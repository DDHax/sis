package store

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path"
)

//原始文件保存目录名
const sourceDirName = "src"

type localStore struct {
}

func (s localStore) md5ToPath(md5Code string) (path string) {
	var buf bytes.Buffer
	buf.WriteString(imagePath)
	buf.WriteByte(os.PathSeparator)
	for _, item := range md5Code {
		buf.WriteRune(item)
		buf.WriteByte(os.PathSeparator)
	}
	return buf.String()
}

func (s localStore) getSrcPath(md5 string) string {
	var buf bytes.Buffer
	buf.WriteString(s.md5ToPath(md5))
	buf.WriteString(sourceDirName)
	buf.WriteByte(os.PathSeparator)
	return buf.String()
}

func (s localStore) write(f multipart.File, md5 string, name string) error {
	//创建目录
	srcPath := s.getSrcPath(md5)
	err := os.MkdirAll(srcPath, os.ModePerm)
	if err != nil {
		return err
	}

	//创建文件
	destFile, err := os.Create(srcPath + name)
	if err != nil {
		return err
	}
	defer destFile.Close()

	//此处必须Seek回起点，否则copy不到东西
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	//文件落地
	_, err = io.Copy(destFile, f)

	return nil
}

func (s localStore) getDirFirstFile(dir string) (string, error) {
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

func (s localStore) read(md5Code string, fileName *string) ([]byte, error) {

	//获取文件路径
	filePath := s.getSrcPath(md5Code)
	if *fileName == "" {
		var err error
		filePath, err = s.getDirFirstFile(filePath)
		if err != nil {
			return nil, err
		}
		*fileName = path.Base(filePath)
	} else {
		filePath = filePath + *fileName
	}

	//读取文件
	return ioutil.ReadFile(filePath)
}
