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
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DDHax/sis/graphics"
	"github.com/DDHax/sis/store"
)

//文件大小上限，此参数将设置为接收图片时分配的内存上限
const maxFileSize = 1024 * 1024 * 50

//文件名长度上限
const maxFileNameLength = 50

//定义缩放图片的极限尺寸
const (
	maxWidth  = 1024 * 1024
	minWidth  = 5
	maxHeight = 1024 * 1024
	minHeight = 5
)

var zeroTime time.Time

func saveFile(f multipart.File, fileName string) (md5Code string, err error) {
	//计算文件MD5
	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	ret := h.Sum(nil)

	//16进制md5转字符串格式
	md5Code = hex.EncodeToString(ret)

	err = store.Write(f, md5Code, fileName)
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

	//读取文件
	md5Code := req.FormValue("md5")
	var fileName string
	data, err := store.Read(md5Code, &fileName)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		return
	}

	//回复文件
	http.ServeContent(w, req, fileName, zeroTime, bytes.NewReader(data))
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
	case ".jpg", ".jpeg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 100})
	case ".png":
		err = png.Encode(&buf, dst)
	case ".gif":
		err = gif.Encode(&buf, dst, nil)
	default:
		log.Print(imgType)
		err = errors.New("找不到编码器")
	}
	return nil, err
}

func stretchSimpleDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//检测参数合法性
	md5Code := req.FormValue("md5")
	height := req.FormValue("h")
	width := req.FormValue("w")
	intW, intH, ret := checkParam(width, height)
	if !ret {
		w.WriteHeader(404)
		return
	}

	//获取原始文件
	var fileName string
	data, err := store.Read(md5Code, &fileName)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		return
	}

	//图像缩放
	dst, err := scaleImage(data, intW, intH)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		return
	}

	//回复文件
	http.ServeContent(w, req, fileName, zeroTime, bytes.NewReader(dst))
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

	data, err := store.Read(md5Code, &fileName)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	//回复文件
	http.ServeContent(w, req, fileName, zeroTime, bytes.NewReader(data))
}

func stretchFullDownHandler(w http.ResponseWriter, req *http.Request) {
	//参数解释
	req.ParseForm()

	//取参
	md5Code := req.FormValue("md5")
	fileName := req.FormValue("file_name")
	height := req.FormValue("h")
	width := req.FormValue("w")
	if !checkFileName(fileName) {
		w.WriteHeader(404)
		return
	}
	intW, intH, ret := checkParam(width, height)
	if !ret {
		w.WriteHeader(404)
		return
	}

	//获取原始文件
	data, err := store.Read(md5Code, &fileName)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		return
	}

	//图像缩放
	dst, err := scaleImage(data, intW, intH)
	if err != nil {
		log.Print(err)
		w.WriteHeader(404)
		return
	}

	//回复文件
	http.ServeContent(w, req, fileName, zeroTime, bytes.NewReader(dst))
}

func defaultHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "./test/upload.html")
}

func main() {
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/up", uploadHandler)
	http.HandleFunc("/simple_down", simpleDownHandler)
	http.HandleFunc("/full_down", fullDownHandler)
	http.HandleFunc("/stretch_simple_down", stretchSimpleDownHandler)
	http.HandleFunc("/stretch_full_down", stretchFullDownHandler)

	//参数解释
	port := flag.String("port", "3333", "监听端口")
	storeType := flag.Bool("localStore", true, "存储类型,true为本地存储，false为远程存储")
	imagePath := flag.String("image", "image", "本地存储时表示本地目录，远程存储时表示远程主机地址")
	flag.Parse()

	store.Init(*imagePath, *storeType)

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
