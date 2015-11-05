package main

import (
	authlib "github.com/clawio/service.auth/lib"
	"github.com/clawio/service.localstore.data/lib"
	"github.com/rs/xlog"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"syscall"
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
	p := lib.MustFromContext(ctx)

	pp := s.getPhysicalPath(p)

	log.Infof("physical path is %s", pp)

	tmpFn, tmpFile, err := s.tmpFile()
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
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
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("copied r.Body into tmp file %s", tmpFn)

	err = tmpFile.Close()
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("closed tmp file %s", tmpFn)

	if err = os.Rename(tmpFn, pp); err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("renamed tmp file %s to %s", tmpFn, pp)

	w.WriteHeader(http.StatusCreated)
}

func (s *server) download(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	log := xlog.FromContext(ctx)
	p := lib.MustFromContext(ctx)

	pp := s.getPhysicalPath(p)

	log.Info("physical path is %s", pp)

	fd, err := os.Open(pp)
	if err == syscall.ENOENT {
		log.Error(err.Error())
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("opened %s", pp)

	info, err := fd.Stat()
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("stated %s got size %d", pp, info.Size())

	if info.IsDir() {
		log.Errorf("%s is a directory", pp)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	defer fd.Close()

	_, err = io.Copy(w, fd)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("copied %s to res.body", pp)

}

func (s *server) authHandler(ctx context.Context, w http.ResponseWriter, r *http.Request,
	next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {

	log := xlog.FromContext(ctx)

	idt, err := s.getIdentityFromReq(r)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	ctx = authlib.NewContext(ctx, idt)
	next(ctx, w, r)
}

func (s *server) accessHandler(ctx context.Context, w http.ResponseWriter, r *http.Request,
	next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {

	log := xlog.FromContext(ctx)
	idt := authlib.MustFromContext(ctx)

	p := getPathFromReq(r) // already sanitized

	if !isUnderHome(p, idt) {
		// TODO use here share service
		log.Warn("access denied to %s accessing %s", *idt, p)
		http.Error(w, "", http.StatusForbidden)
		return
	}

	if p == getHome(idt) {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	log.Infof("path is %s", p)

	ctx = lib.NewContext(ctx, p)
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

func (s *server) getPhysicalPath(p string) string {
	return path.Join(s.p.dataDir, path.Clean(p))
}

func (s *server) getIdentityFromReq(r *http.Request) (*authlib.Identity, error) {

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

	return authlib.ParseToken(token, s.p.sharedSecret)
}
