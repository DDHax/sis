package client

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"testing"
)

type clientTest struct {
	fileName string
	response string
	md5      string
}

var clientTests = []clientTest{
	{
		fileName: "test1.jpg",
		md5:      "685264ff36effb53d7ecdb81d3b89b22",
	},
	{
		fileName: "测试2.png",
		md5:      "6b602ffddcc45c254217168a98420153",
	},
	{
		fileName: "test3.gif",
		md5:      "b8602cf392b801d60281681e56299f17",
	},
}

func Test_singleUpload(t *testing.T) {
	rep, err := singleUpload(clientTests[0].fileName)
	if err != nil {
		t.Error(err)
	}

	type Message struct {
		Name, MD5 string
	}
	dec := json.NewDecoder(strings.NewReader(rep))

	// read open bracket
	_, err = dec.Token()
	if err != nil {
		t.Fatal(err)
	}
	for dec.More() {
		var m Message
		// decode an array value (Message)
		err := dec.Decode(&m)
		if err != nil {
			t.Fatal(err)
		}

		if m.Name != clientTests[0].fileName || m.MD5 != clientTests[0].md5 {
			t.Errorf("非预期返回值, 预期[%s][%s] , 实际[%s][%s]", clientTests[0].fileName, clientTests[0].md5, m.Name, m.MD5)
		}
	}

}

func Test_multipleUpload(t *testing.T) {
	var files = []string{clientTests[1].fileName, clientTests[2].fileName}
	rep, err := multipleUpload(files)
	if err != nil {
		t.Error(err)
	}

	type Message struct {
		Name, MD5 string
	}
	dec := json.NewDecoder(strings.NewReader(rep))

	// read open bracket
	_, err = dec.Token()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < 3; i++ {
		dec.More()
		var m Message
		// decode an array value (Message)
		err := dec.Decode(&m)
		if err != nil {
			t.Fatal(err)
		}

		if m.Name != clientTests[i].fileName || m.MD5 != clientTests[i].md5 {
			t.Errorf("非预期返回值, 预期[%s][%s] , 实际[%s][%s]", clientTests[i].fileName, clientTests[i].md5, m.Name, m.MD5)
		}
	}
}

func Test_simpleDown(t *testing.T) {
	buf, err := simpleDown(clientTests[0].md5)
	if err != nil {
		t.Fatal(err)
	}
	srcCode := md5.Sum(buf)
	md5Code := hex.EncodeToString(srcCode[:])
	if md5Code != clientTests[0].md5 {
		t.Fatal(err)
	}
}

func Test_stretchSimpleDown(t *testing.T) {
	const w = 200
	const h = 100
	data, err := stretchSimpleDown(clientTests[0].md5, w, h)
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(data)

	img, _, err := image.Decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	rect := img.Bounds()
	if rect.Dx() != w || rect.Dy() != h {
		t.Fatal(err)
	}
}

func Test_fullDown(t *testing.T) {
	buf, err := fullDown(clientTests[1].md5, clientTests[1].fileName)
	if err != nil {
		t.Fatal(err)
	}
	srcCode := md5.Sum(buf)
	md5Code := hex.EncodeToString(srcCode[:])
	if md5Code != clientTests[1].md5 {
		t.Fatal(err)
	}
}

func Test_stretchFullDown(t *testing.T) {
	const w = 300
	const h = 200
	data, err := stretchFullDown(clientTests[1].md5, clientTests[1].fileName, w, h)
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewBuffer(data)

	img, _, err := image.Decode(buf)
	if err != nil {
		t.Fatal(err)
	}
	rect := img.Bounds()
	if rect.Dx() != w || rect.Dy() != h {
		t.Fatal(err)
	}
}
