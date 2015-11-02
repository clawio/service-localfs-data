package main

import (
	"github.com/clawio/service.auth/lib"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

const (
	dirPerm = 0755
)

type newServerParams struct {
	dataDir      string
	tmpDir       string
	sharedSecret string
}

func newServer(p *newServerParams) *server {
	return &server{p}
}

type server struct {
	p *newServerParams
}

func (s *server) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if strings.ToUpper(r.Method) == "PUT" {
		s.authHandler(ctx, w, r, s.upload)
	} else if strings.ToUpper(r.Method) == "GET" {
		s.authHandler(ctx, w, r, s.download)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (s *server) upload(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	log := xlog.FromContext(ctx)
	idt := lib.MustFromContext(ctx)

	p := s.getFilePath(r, idt)

	tmpFn, tmpFile, err := s.tmpFile()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("created tmp file %s", tmpFn)

	// TODO(labkode) Sometimes ContentLength = -1 because it is a binary
	// upload with TransferEncoding: chunked.
	// Instead using Copy we shoudl use a LimitedReader with a max file upload
	// configuration value.
	_, err = io.Copy(tmpFile, r.Body)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("copied r.body into tmp file %s", tmpFn)

	err = tmpFile.Close()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("closed tmp file %s", tmpFn)

	if err = os.Rename(tmpFn, p); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Infof("renamed tmp file %s to %s", tmpFn, p)

	w.WriteHeader(http.StatusCreated)
}

func (s *server) download(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	log := xlog.FromContext(ctx)
	idt := lib.MustFromContext(ctx)

	p := s.getFilePath(r, idt)

	fd, err := os.Open(p)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer fd.Close()

	_, err = io.Copy(w, fd)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func (s *server) authHandler(ctx context.Context, w http.ResponseWriter, r *http.Request,
	next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {

	log := xlog.FromContext(ctx)

	idt, err := s.getIdentityFromReq(r)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ctx = lib.NewContext(ctx, idt)
	next(ctx, w, r)
}

func (s *server) tmpFile() (string, *os.File, error) {

	file, err := ioutil.TempFile(s.p.tmpDir, serviceID)
	if err != nil {
		return "", nil, err
	}

	fn := path.Join(path.Clean(file.Name()))

	return fn, file, nil
}
