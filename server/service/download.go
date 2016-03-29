package service

import (
	"io"
	"net/http"
	"os"

	"github.com/NYTimes/gizmo/server"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/service-auth/server/spec"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Download streams a file to the client.
func (s *Service) Download(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	identity := context.Get(r, identityKey).(*spec.Identity)
	storagePath := s.getStoragePath(identity, path)
	fd, err := os.Open(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error opening file")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, fd); err != nil {
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error writing response body")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}
