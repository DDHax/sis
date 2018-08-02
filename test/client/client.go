package client

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	urlUp = "http://127.0.0.1:3333/up"
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
	for i := 0; i < 2; i++ {
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
