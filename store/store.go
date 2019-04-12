package store

import (
	"mime/multipart"
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
	if err != nil && gCache.maxSize != 0 {
		gCache.write(f, md5, name)
	}
	return err
}

//Read 读取图像文件接口
func Read(md5Code string, fileName *string) ([]byte, error) {
	//读取缓存
	if gCache.maxSize != 0 {
		data, err := gCache.read(md5Code, fileName)
		if err == nil {
			return data, err
		}
	}

	//读原始文件
	return storer.read(md5Code, fileName)
}

//Init 初始化接口，设置存储路径和类型
func Init(path string, isLocal bool, cacheSize int) {
	imagePath = path
	gCache.maxSize = int64(cacheSize) * 1024 * 1024
	if isLocal {
		storer, _ = storer.(localStore)
	} else {
		storer, _ = storer.(remoteStore)
	}
}
