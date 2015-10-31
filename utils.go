package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func getTokenFromReq(r *http.Request) (string, error) {

	// Look for an Authorization header
	if ah := r.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:6]) == "BEARER" {
			return ah[7:], nil
		}
	}

	// Look for "auth_token" parameter
	r.ParseMultipartForm(10e6)
	if tokStr := r.Form.Get("auth_token"); tokStr != "" {
		return tokStr, nil
	}

	return "", fmt.Errorf("no auth token in req")
}

func getReqPath(r *http.Request) string {

	if len(r.URL.Path) > len(endPoint) {
		return strings.TrimPrefix(r.URL.Path[len(endPoint):], "/")
	}
	return ""
}

func (s *server) getFilePath(r *http.Request) string {
	return path.Join(s.p.dataDir, path.Clean(getReqPath(r)))
}

func copyFile(src, dst string, size int64) (err error) {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.CopyN(writer, reader, size)
	if err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) (err error) {
	err = os.Mkdir(dst, dirPerm)
	if err != nil {
		return err
	}

	directory, err := os.Open(src)
	if err != nil {
		return err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		_src := path.Join(src, obj.Name())
		_dst := path.Join(dst, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(_src, _dst)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = copyFile(_src, _dst, obj.Size())
			if err != nil {
				return err
			}
		}
	}
	return
}
