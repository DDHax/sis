package store

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"mime/multipart"
	"strconv"

	"github.com/DDHax/sis/store/graphics"
)

//FileIO 文件读写接口
type fileIO interface {
	write(f multipart.File, md5 string, name string) error
	read(md5Code string, fileName *string) ([]byte, error)
}

var imagePath string
var storer fileIO

//Write 写入图像文件接口
func Write(f multipart.File, md5 string, name string) error {
	//落地写入
	err := storer.write(f, md5, name)

	//写入缓存
	if err == nil && gCache.isEnable() {
		//log.Printf("写入cache %v %v", md5, name)
		gCache.write(f, md5, name)
	}
	return err
}

func scaleImage(data []byte, destW, destH int) ([]byte, error) {
	//解码原始图像
	img, imgType, err := image.Decode(bytes.NewReader(data))

	//建立目标图形
	dst := image.NewRGBA(image.Rect(0, 0, destW, destH))

	//执行缩放
	err = graphics.Scale(dst, img)
	if err != nil {
		return nil, err
	}

	//编码缩放后图像
	var buf bytes.Buffer
	switch imgType {
	case "jpg", "jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 100})
	case "png":
		err = png.Encode(&buf, dst)
	case "gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		log.Print(imgType)
		err = errors.New("找不到编码器")
	}
	return buf.Bytes(), err
}

//Read 读取图像文件接口
func Read(md5Code string, fileName *string, width, height int) ([]byte, error) {

	key := md5Code + *fileName
	longKey := key
	if width > 0 && height > 0 {
		longKey = longKey + strconv.Itoa(width) + "_" + strconv.Itoa(height)
	}

	//读取缓存
	if gCache.isEnable() {
		data, err := gCache.read(longKey)
		if err == nil {
			//log.Printf("命中cache %v %v", md5Code, fileName)
			return data, err
		}
	}

	//读原始文件
	data, err := storer.read(md5Code, fileName)
	if err == nil && width > 0 && height > 0 {
		//图像缩放
		dst, err := scaleImage(data, width, height)
		if err == nil && gCache.isEnable() {
			//写入缓存
			gCache.memWrite(longKey, dst)
		}
		return dst, err
	}

	return data, err
}

//Init 初始化接口，设置存储路径和类型
func Init(path string, isLocal bool, cacheSize int) {
	imagePath = path
	gCache.maxSize = int64(cacheSize) * 1024 * 1024
	gCache.data = make(map[string][]byte)
	if isLocal {
		storer, _ = storer.(localStore)
	} else {
		storer, _ = storer.(remoteStore)
	}
}
