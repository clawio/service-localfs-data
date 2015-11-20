package main

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	authlib "github.com/clawio/service.auth/lib"
	"github.com/clawio/service.localstore.data/lib"
	pb "github.com/clawio/service.localstore.data/proto/propagator"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"hash"
	"hash/adler32"
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
	checksum     string
	prop         string
	sharedSecret string
}

func newServer(p *newServerParams) (*server, error) {

	s := &server{}
	s.p = p

	return s, nil
}

type server struct {
	p *newServerParams
}

func (s *server) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	reqLogger := log.WithField("trace", getTraceID(r))
	ctx = NewLogContext(ctx, reqLogger)

	reqLogger.WithField("url", r.URL.String())

	if strings.ToUpper(r.Method) == "PUT" {
		reqLogger.WithField("op", "upload")
		s.authHandler(ctx, w, r, s.upload)
	} else if strings.ToUpper(r.Method) == "GET" {
		reqLogger.WithField("op", "download")
		s.authHandler(ctx, w, r, s.download)
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (s *server) upload(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	log := MustFromLogContext(ctx)
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

	var mw io.Writer
	var hasher hash.Hash
	var isChecksumed bool
	var computedChecksum string

	switch s.p.checksum {
	case "md5":
		hasher = md5.New()
		isChecksumed = true
		mw = io.MultiWriter(tmpFile, hasher)
	case "sha1":
		hasher = sha1.New()
		isChecksumed = true
		mw = io.MultiWriter(tmpFile, hasher)
	case "adler32":
		hasher = adler32.New()
		isChecksumed = true
		mw = io.MultiWriter(tmpFile, hasher)
	default:
		mw = io.MultiWriter(tmpFile)
	}

	// TODO(labkode) Sometimes ContentLength = -1 because it is a binary
	// upload with TransferEncoding: chunked.
	// Instead using Copy we shoudl use a LimitedReader with a max file upload
	// configuration value.
	_, err = io.Copy(mw, r.Body)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	clientSumType, clientSum := s.getChecksumInfo(r)

	if isChecksumed {
		log.Infof("file sent with checksum %s", clientSumType+":"+clientSum)

		// checksums are given in hexadecimal format.
		computedChecksum = fmt.Sprintf("%x", string(hasher.Sum(nil)))

		if clientSumType == s.p.checksum && clientSum != "" {

			isCorrupted := computedChecksum != clientSum

			if isCorrupted {
				log.Errorf("corrupted file. expected %s and got %s",
					s.p.checksum+":"+computedChecksum, clientSumType+":"+clientSum)
				http.Error(w, "", http.StatusPreconditionFailed)
				return
			}
		}
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

	con, err := grpc.Dial(s.p.prop, grpc.WithInsecure())
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer con.Close()

	log.Infof("created connection to prop")

	client := pb.NewPropClient(con)

	in := &pb.PutReq{}
	in.Path = p
	in.AccessToken = authlib.MustFromTokenContext(ctx)
	in.Checksum = clientSumType + ":" + clientSum

	_, err = client.Put(ctx, in)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	log.Infof("putted path %s into prop", p)

	w.WriteHeader(http.StatusCreated)
}

func (s *server) download(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	log := MustFromLogContext(ctx)
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

	log := MustFromLogContext(ctx)

	idt, err := s.getIdentityFromReq(r)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	p := getPathFromReq(r) // already sanitized

	if !isUnderHome(p, idt) {
		// TODO use here share service
		log.Warnf("%s cannot access %s", *idt, p)
		http.Error(w, "", http.StatusForbidden)
		return
	}

	if p == getHome(idt) {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	log.Infof("path is %s", p)

	ctx = authlib.NewContext(ctx, idt)
	ctx = lib.NewContext(ctx, p)
	ctx = authlib.NewTokenContext(ctx, s.getTokenFromReq(r))
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

func (s *server) getTokenFromReq(r *http.Request) string {

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

	return token
}

func (s *server) getIdentityFromReq(r *http.Request) (*authlib.Identity, error) {
	return authlib.ParseToken(s.getTokenFromReq(r), s.p.sharedSecret)
}
