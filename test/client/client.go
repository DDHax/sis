package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	urlUp                = "http://127.0.0.1:3333/up"
	urlSimpleDown        = "http://127.0.0.1:3333/simple_down?md5=%s"
	urlStretchSimpleDown = "http://127.0.0.1:3333/stretch_simple_down?md5=%s&w=%d&h=%d"
	urlFullDown          = "http://127.0.0.1:3333/full_down?md5=%s&file_name=%s"
	urlStretchFullDown   = "http://127.0.0.1:3333/stretch_full_down?md5=%s&file_name=%s&w=%d&h=%d"
)

func singleUpload(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	mt := multipart.NewWriter(&buf)
	fileWriter, err := mt.CreateFormFile("upload_test", fileName)
	if err != nil {
		return "", err
	}
	fileReader := bufio.NewReader(file)
	fileReader.WriteTo(fileWriter)
	mt.Close()

	contentType := "multipart/form-data;boundary=" + mt.Boundary()
	req, err := http.NewRequest("POST", urlUp, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	content := string(body)
	return content, nil
}

func multipleUpload(files []string) (string, error) {

	var buf bytes.Buffer
	mt := multipart.NewWriter(&buf)
	for i := 0; i < len(files); i++ {
		file, err := os.Open(files[i])
		if err != nil {
			return "", err
		}
		defer file.Close()

		fileWriter, err := mt.CreateFormFile("upload_test", files[i])
		if err != nil {
			return "", err
		}
		fileReader := bufio.NewReader(file)
		fileReader.WriteTo(fileWriter)
	}
	mt.Close()

	contentType := "multipart/form-data;boundary=" + mt.Boundary()
	req, err := http.NewRequest("POST", urlUp, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	content := string(body)
	return content, nil
}

func simpleDown(md5 string) ([]byte, error) {
	url := fmt.Sprintf(urlSimpleDown, md5)
	resp, err := http.Get(url)
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

func stretchSimpleDown(md5 string, w, h int) ([]byte, error) {
	url := fmt.Sprintf(urlStretchSimpleDown, md5, w, h)
	resp, err := http.Get(url)
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

func fullDown(md5, fileName string) ([]byte, error) {
	url := fmt.Sprintf(urlFullDown, md5, fileName)
	resp, err := http.Get(url)
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

func stretchFullDown(md5, fileName string, w, h int) ([]byte, error) {
	url := fmt.Sprintf(urlStretchFullDown, md5, fileName, w, h)
	resp, err := http.Get(url)
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
