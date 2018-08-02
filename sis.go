package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/DDHax/sis/graphics"
)

//文件大小上限，此参数将设置为接收图片时分配的内存上限
const maxFileSize = 1024 * 1024 * 50

//文件名长度上限
const maxFileNameLength = 50

//原始文件保存目录名
const sourceDirName = "src"

//定义缩放图片的极限尺寸
const (
	maxWidth  = 1024 * 1024
	minWidth  = 5
	maxHeight = 1024 * 1024
	minHeight = 5
)

var imagePath string

func md5ToPath(md5Code string) (path string) {
	var buf bytes.Buffer
	buf.WriteString(imagePath)
	buf.WriteByte(os.PathSeparator)
	for _, item := range md5Code {
		buf.WriteRune(item)
		buf.WriteByte(os.PathSeparator)
	}
	return buf.String()
}

func getSrcPath(md5 string) string {
	var buf bytes.Buffer
	buf.WriteString(md5ToPath(md5))
	buf.WriteString(sourceDirName)
	buf.WriteByte(os.PathSeparator)
	return buf.String()
}

func getStretchPath(md5, width, height string) string {
	var buf bytes.Buffer
	buf.WriteString(md5ToPath(md5))
	buf.WriteString(width)
	buf.WriteByte('_')
	buf.WriteString(height)
	buf.WriteByte(os.PathSeparator)
	return buf.String()
}

func getDirFirstFile(dir string) (string, error) {
	//获取目录信息
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	//查找第一个文件返回路径
	for _, file := range files {
		return dir + file.Name(), nil
	}
	return "", errors.New("目录中没有文件")
}

func saveFile(f multipart.File, fileName string) (md5Code string, err error) {
	//计算文件MD5
	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	ret := h.Sum(nil)

	//16进制md5转字符串格式
	md5Code = hex.EncodeToString(ret)

	//创建目录
	srcPath := getSrcPath(md5Code)
	err = os.MkdirAll(srcPath, os.ModeType)
	if err != nil {
		return
	}

	//创建文件
	destFile, err := os.Create(srcPath + fileName)
	if err != nil {
		return
	}
	defer destFile.Close()

	//此处必须Seek回起点，否则copy不到东西
	_, err = f.Seek(0, 0)
	if err != nil {
		return
	}

	//文件落地
	_, err = io.Copy(destFile, f)

	return
}

//检测文件名合法性,包括长度和安全性检测
func checkFileName(inputFileName string) bool {
	if len(inputFileName) > maxFileNameLength ||
		inputFileName != filepath.Base(inputFileName) {
		log.Printf("收到非预期文件：%s", inputFileName)
		return false
	}
	return true
}

func uploadHandler(w http.ResponseWriter, req *http.Request) {
	//这个必须得有，客户端问的时候总要回答一下，否则测试页面无法工作
	if strings.ToUpper(req.Method) == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Method", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(204)
		return
	}

	//响应
	status := 400
	message := "不要乱来"
	defer func(w http.ResponseWriter) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(status)
		w.Write([]byte(message))
	}(w)

	//处理上传请求
	if strings.ToUpper(req.Method) == "POST" {
		//文件大小检查
		length, _ := strconv.Atoi(req.Header.Get("Content-Length"))
		if length > maxFileSize {
			status = 413
			message = "上传文件超出50M限制"
			return
		}

		//解释请求
		err := req.ParseMultipartForm(int64(length))
		if err != nil {
			log.Print(err)
			return
		}

		//处理请求
		if req.MultipartForm != nil {
			//初始化返回值
			var messageBuf bytes.Buffer
			messageBuf.WriteString(`[`)
			for _, v := range req.MultipartForm.File {
				for _, fileHead := range v {

					//文件名长度检查
					if !checkFileName(fileHead.Filename) {
						status = 413
						message = "文件名长度超出50字节限制"
						return
					}

					//将上传的文件buffer转为文件接口
					file, err := fileHead.Open()
					if err != nil {
						status = 500
						message = "发生诡异错误"
						return
					}

					//保存文件
					md5Code, err := saveFile(file, fileHead.Filename)
					if err != nil {
						status = 500
						message = "创建文件失败"
						return
					}

					//写入json格式返回值
					if messageBuf.Len() > 1 {
						messageBuf.WriteByte(',')
					}
					messageBuf.WriteString(`{"Name":"`)
					messageBuf.WriteString(fileHead.Filename)
					messageBuf.WriteString(`", "MD5":"`)
					messageBuf.WriteString(md5Code)
					messageBuf.WriteString(`"}`)
				}
			}
			messageBuf.WriteString("]")
			status = 200
			message = messageBuf.String()
		}
	}
}

func simpleDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//定位目录
	md5Code := req.FormValue("md5")

	//获取文件路径
	filePath, err := getDirFirstFile(getSrcPath(md5Code))
	if err != nil {
		w.WriteHeader(404)
		return
	} else {
		http.ServeFile(w, req, filePath)
		return
	}
}

func checkParam(w, h string) (int, int, bool) {
	intW, err := strconv.Atoi(w)
	if err != nil {
		return 0, 0, false
	}
	intH, err := strconv.Atoi(h)
	if err != nil {
		return 0, 0, false
	}
	if intW < minWidth || intW > maxWidth || intH < minHeight || intH > maxHeight {
		return 0, 0, false
	}
	return intW, intH, true
}

func loadImage(path string) (img image.Image, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	img, _, err = image.Decode(file)
	return
}

func saveImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	ext := filepath.Ext(path)
	switch ext {
	case "jpg", "jpeg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	case "png":
		err = png.Encode(file, img)
	case "gif":
		err = gif.Encode(file, img, nil)
	default:
		err = errors.New("找不到编码器")
	}
	return err
}

func generateFile(md5, stretchPath, fileName string, w, h int) (string, error) {
	//找到原始文件
	var srcFilePath string
	var err error
	srcPath := getSrcPath(md5)
	if fileName != "" {
		srcFilePath = srcPath + fileName
	} else {
		srcFilePath, err = getDirFirstFile(srcPath)
		if err != nil {
			return "", err
		}
	}

	//载入原始文件
	src, err := loadImage(srcFilePath)
	if err != nil {
		return "", err
	}

	//建立目标图形
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	//执行缩放
	err = graphics.Scale(dst, src)
	if err != nil {
		return "", err
	}

	//创建目标目录
	err = os.MkdirAll(stretchPath, os.ModeType)
	if err != nil {
		return "", err
	}

	//目标图形存盘
	dstFilePath := stretchPath + filepath.Base(srcFilePath)
	err = saveImage(dstFilePath, dst)
	if err != nil {
		return "", err
	}
	return dstFilePath, nil
}

func stretchSimpleDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//检测参数合法性
	md5Code := req.FormValue("md5")
	width := req.FormValue("h")
	height := req.FormValue("w")
	intW, intH, ret := checkParam(width, height)
	if !ret {
		w.WriteHeader(404)
		return
	}

	//定位目录
	stretchPath := getStretchPath(md5Code, width, height)

	//查找目录中第一个文件
	stretchFullPath, err := getDirFirstFile(stretchPath)
	if err != nil {
		//目录或文件不存在，创建文件
		stretchFullPath, err = generateFile(md5Code, stretchPath, "", intW, intH)
		if err != nil {
			w.WriteHeader(500)
			return
		} else {
			http.ServeFile(w, req, stretchFullPath)
			return
		}
	} else {
		//找到缩放文件，直接返回
		http.ServeFile(w, req, stretchFullPath)
		return
	}
}

func fullDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//定位目录
	md5Code := req.FormValue("md5")
	fileName := req.FormValue("file_name")
	if !checkFileName(fileName) {
		w.WriteHeader(404)
		return
	}

	filePath := getSrcPath(md5Code)

	http.ServeFile(w, req, filePath+fileName)
	return
}

func stretchFullDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//取参
	md5Code := req.FormValue("md5")
	fileName := req.FormValue("file_name")
	width := req.FormValue("h")
	height := req.FormValue("w")
	if !checkFileName(fileName) {
		w.WriteHeader(404)
		return
	}
	intW, intH, ret := checkParam(width, height)
	if !ret {
		w.WriteHeader(404)
		return
	}

	//定位路径
	stretchPath := getStretchPath(md5Code, width, height)
	stretchFullPath := stretchPath + fileName

	//判断文件是否存在
	if _, err := os.Stat(stretchFullPath); os.IsNotExist(err) {
		//文件不存在，创建文件
		_, err = generateFile(md5Code, stretchPath, fileName, intW, intH)
		if err != nil {
			w.WriteHeader(500)
			return
		} else {
			http.ServeFile(w, req, stretchFullPath)
			return
		}
	} else {
		//缩放文件已存在，直接返回
		http.ServeFile(w, req, stretchFullPath)
		return
	}
}

func main() {
	http.HandleFunc("/up", uploadHandler)
	http.HandleFunc("/simple_down", simpleDownHandler)
	http.HandleFunc("/full_down", fullDownHandler)
	http.HandleFunc("/stretch_simple_down", stretchSimpleDownHandler)
	http.HandleFunc("/stretch_full_down", stretchFullDownHandler)

	var port = flag.String("port", "3333", "监听端口")
	flag.StringVar(&imagePath, "image", "image", "图片存储目录")

	flag.Parse()

	var srv http.Server
	srv.Addr = ":" + *port

	//下面实现HTTP服务优雅退出，代码摘自官方文档
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
		)

		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
}
