package main

import (
	"github.com/clawio/service.auth/lib"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func (s *server) getIdentityFromReq(r *http.Request) (*lib.Identity, error) {

	var token string

	// Look for an Authorization header
	if ah := r.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:6]) == "BEARER" {
			token = ah[7:]
		}
	}

	if token == "" {
		// Look for "auth_token" parameter
		r.ParseMultipartForm(10e6)
		if tokStr := r.Form.Get("access_token"); tokStr != "" {
			token = tokStr
		}

	}

	return lib.ParseToken(token, s.p.sharedSecret)
}

func getReqPath(r *http.Request) string {

	if len(r.URL.Path) > len(endPoint) {
		return strings.TrimPrefix(r.URL.Path[len(endPoint):], "/")
	}
	return ""
}

func (s *server) getFilePath(r *http.Request, idt *lib.Identity) string {
	return path.Join(s.getHome(idt), path.Clean(getReqPath(r)))
}

func (s *server) getHome(idt *lib.Identity) string {
	return path.Join(s.p.dataDir, path.Join(idt.Pid))
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
