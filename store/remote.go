package store

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

const (
	urlDerectUp   = "/derect_up"
	urlSimpleDown = "/simple_down?md5=%s"
	urlFullDown   = "/full_down?md5=%s&file_name=%s"
)

type remoteStore struct {
}

func (r remoteStore) write(f multipart.File, md5 string, name string) error {
	var buf bytes.Buffer
	mt := multipart.NewWriter(&buf)
	fileWriter, err := mt.CreateFormFile(md5, name)
	if err != nil {
		return err
	}
	f.Seek(0, 0)
	fileReader := bufio.NewReader(f)
	fileReader.WriteTo(fileWriter)
	mt.Close()

	contentType := "multipart/form-data;boundary=" + mt.Boundary()
	url := imagePath + urlDerectUp
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = errors.New(resp.Status)
	}
	return err
}

func (r remoteStore) read(md5Code string, fileName *string) ([]byte, error) {
	url := fmt.Sprintf(urlSimpleDown, md5Code)
	if *fileName != "" {
		url = fmt.Sprintf(urlFullDown, md5Code, *fileName)
	}

	resp, err := http.Get(imagePath + url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
