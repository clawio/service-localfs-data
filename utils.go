package main

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/clawio/service.auth/lib"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

func getPathFromReq(r *http.Request) string {
	return path.Clean(r.URL.Path)
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

func newGRPCTraceContext(ctx context.Context, trace string) context.Context {
	md := metadata.Pairs("trace", trace)
	ctx = metadata.NewContext(ctx, md)
	return ctx
}

type checksum struct {
	Type string
	Sum  string
}

func (c *checksum) String() string {
	if c.Type == "" {
		return ""
	}
	return c.Type + ":" + c.Sum
}

// getChecksumInfo retrieves checksum information sent by a client via query params or via header.
// If the checksum is sent in the header the header must be called X-Checksum and the content must be:
// <checksumtype>:<checksum>.
// If the info is sent in the URL the name of the query param is checksum and thas the same format
// as in the header.
func (a *server) getChecksumInfo(r *http.Request) *checksum {

	var checksumInfo string
	var checksumType string
	var sum string

	// 1. Get checksum info from query params
	checksumInfo = r.URL.Query().Get("checksum")
	if checksumInfo != "" {
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			sum = parts[1]
		}
	}

	// 2. Get checksum info from header
	if checksumInfo == "" { // If already provided in URL we don´t override
		checksumInfo = r.Header.Get("CIO-Checksum")
		parts := strings.Split(checksumInfo, ":")
		if len(parts) > 1 {
			checksumType = parts[0]
			sum = parts[1]
		}
	}

	return &checksum{checksumType, sum}
}

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// logKey is the context key for an identity.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const logKey key = 0

// NewLogContext returns a new Context carrying a logger.
func NewLogContext(ctx context.Context, logger *log.Entry) context.Context {
	return context.WithValue(ctx, logKey, logger)

}

// MustFromLogContext extracts the logger from ctx.
// If not present it panics.
func MustFromLogContext(ctx context.Context) *log.Entry {
	val, ok := ctx.Value(logKey).(*log.Entry)
	if !ok {
		panic("logger is not registered")

	}
	return val

}
