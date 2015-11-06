package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/clawio/service.auth/lib"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func getPathFromReq(r *http.Request) string {

	if len(r.URL.Path) > len(endPoint) {
		return path.Clean(r.URL.Path[len(endPoint):])
	}
	return ""
}

// getHome returns the user home directory.
// the logical home has this layout.
// local/users/<letter>/<pid>
// Example: /local/users/o/ourense
// idt.Pid must be always non-empty
func getHome(idt *lib.Identity) string {

	pid := path.Clean(idt.Pid)

	if pid == "" {
		panic("idt.Pid must not be empty")
	}

	return path.Join("/local", "users", string(pid[0]), pid)
}

func isUnderHome(p string, idt *lib.Identity) bool {

	p = path.Clean(p)

	if strings.HasPrefix(p, getHome(idt)) {
		return true
	}

	return false
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

// getTraceID returns the traceID that comes in the request
// or generate a new one
func getTraceID(r *http.Request) string {
	traceID := r.Header.Get("CIO-TraceID")
	if traceID == "" {
		return uuid.New()
	}
	return traceID
}
