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
	return storer.write(f, md5, name)
}

//Read 读取图像文件接口
func Read(md5Code string, fileName *string) ([]byte, error) {
	return storer.read(md5Code, fileName)
}

//Init 初始化接口，设置存储路径和类型
func Init(path string, isLocal bool) {
	imagePath = path
	if isLocal {
		storer, _ = storer.(localStore)
	} else {
		storer, _ = storer.(remoteStore)
	}
}
