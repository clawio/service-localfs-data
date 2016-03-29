package service

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"hash/adler32"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/NYTimes/gizmo/server"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/service-auth/server/spec"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Upload saves a file to disk.
// It has 4 phases:
// 1) Write file to a temporary directory
// 2) Optional: calculate checksum of the file if checksum is configured
// 3) Optional: verify if server-checksum matches client-checksum if provided
// 4) Move the file from temporary directory to user directory. This operation must be atomic.
func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// 1) Write file to a temporary directory
	readCloser := http.MaxBytesReader(w, r.Body, s.Config.Storage.RequestBodyMaxSize)
	tempFileName, err := s.saveToTempFile(readCloser)
	if err != nil {
		if err.Error() == "http: request body too large" {
			server.Log.WithFields(logrus.Fields{
				"error": err,
			}).Warn("request body max size exceed")
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return

		}
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error writing to temp file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// 2) Optional: calculate checksum of the file if checksum is configured
	if s.Config.Storage.Checksum != "" {
		serverChecksum, err := s.calculateChecksumForFile(tempFileName)
		if err != nil {
			// TODO(labkode) check that if error is checksum not supported reply with NotImplemented.
			server.Log.WithFields(logrus.Fields{
				"error": err,
			}).Error("error calculating checksum")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// add serverChecksum to header
		w.Header().Set("checksum", serverChecksum)

		// 3) Optional: verify if server-checksum matches client-checksum if provided
		if s.Config.Storage.VerifyClientChecksum {
			clientChecksum := s.obtainClientChecksum(r)
			if serverChecksum != clientChecksum {
				server.Log.WithFields(logrus.Fields{
					"clientchksum": clientChecksum,
					"serverchksum": serverChecksum,
				}).Warn("client and server checksum do not match")
				http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
				return
			}
		}
	}

	// 4) Move the file from temporary directory to user directory. This operation must be atomic.
	path := mux.Vars(r)["path"]
	identity := context.Get(r, identityKey).(*spec.Identity)
	storagePath := s.getStoragePath(identity, path)
	if err := os.Rename(tempFileName, storagePath); err != nil {
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error renaming file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	return
}

func (s *Service) saveToTempFile(r io.Reader) (string, error) {
	fd, err := ioutil.TempFile(s.Config.Storage.TempDir, "")
	defer fd.Close()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fd, r); err != nil {
		return "", err
	}
	return fd.Name(), nil
}

func (s *Service) calculateChecksumForFile(fn string) (string, error) {
	checksumType := strings.ToLower(s.Config.Storage.Checksum)
	var hash hash.Hash
	switch checksumType {
	case "md5":
		hash = md5.New()
	case "adler32":
		hash = adler32.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	default:
		return "", errors.New("provided checksum not implemented")
	}
	fd, err := os.Open(fn)
	defer fd.Close()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(hash, fd); err != nil {
		return "", err
	}
	checksum := fmt.Sprintf("%x", hash.Sum([]byte{}))
	return checksumType + ":" + checksum, nil
}

func (s *Service) obtainClientChecksum(r *http.Request) string {
	if t := r.Header.Get("checksum"); t != "" {
		return t
	}
	return r.URL.Query().Get("checksum")
}
