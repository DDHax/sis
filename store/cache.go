package store

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
)

type cache struct {
	data    map[string][]byte //图片缓存
	keyList []string          //按照时间顺序排列的图片KEY列表
	useSize int64             //缓存已用空间
	maxSize int64             //最大缓存
}

var gCache cache

func computeSize(keyLen, dataLen int) int {
	//map 中每个项额外占用大概10.79字节，此处取11表示
	const extraLen = 11

	return keyLen*2 + dataLen + extraLen
}

//准备cache空间，注意此函数仅在逻辑上释放已占用空间，真正的内存回收依赖Golang的GC
func (c *cache) prepare(expectLen int) error {

	//请求空间超出能力
	if c.maxSize < int64(expectLen) {
		return errors.New("缓存空间不足")
	}

	//剩余空间充足
	if c.maxSize-c.useSize >= int64(expectLen) {
		return nil
	}

	//释放空间
	var releaseLen int
	var releaseNum int
	for _, key := range c.keyList {
		releaseLen = releaseLen + computeSize(len(key), len(c.data[key]))
		delete(c.data, key)

		releaseNum = releaseNum + 1
		if releaseLen >= expectLen {
			break
		}
	}
	c.keyList = c.keyList[releaseNum:]
	c.useSize = c.useSize - int64(releaseLen)
	return nil
}

func (c *cache) write(f multipart.File, md5 string, name string) error {
	//此处必须Seek回起点，否则copy不到东西
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}

	//文件写入内存
	var buf bytes.Buffer
	_, err = io.Copy(&buf, f)
	if err != nil {
		return err
	}

	//准备缓存空间
	key := md5 + name
	err = c.prepare(computeSize(len(key), buf.Len()))
	if err != nil {
		return err
	}

	//写入缓存
	c.derectWrite(key, buf.Bytes())
	return nil
}

func (c *cache) derectWrite(key string, data []byte) {
	c.data[key] = data
	c.keyList = append(c.keyList, key)
	c.useSize = c.useSize + int64(computeSize(len(key), len(data)))
}

func (c *cache) read(md5Code string, fileName *string) ([]byte, error) {
	if v, ok := c.data[md5Code+*fileName]; ok {
		return v, nil
	}
	return nil, errors.New("缓存未命中")
}

func (c *cache) memWrite(key string, data []byte) error {
	//准备缓存空间
	err := c.prepare(computeSize(len(key), len(data)))
	if err != nil {
		return err
	}

	//写入缓存
	c.derectWrite(key, data)
	return nil
}
