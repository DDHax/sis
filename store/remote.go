package store

import (
	"mime/multipart"
)

type remoteStore struct {
}

func (r remoteStore) write(f multipart.File, md5 string, name string) error {
	return nil
}

func (r remoteStore) read(md5Code string, fileName *string) ([]byte, error) {
	return nil, nil
}
