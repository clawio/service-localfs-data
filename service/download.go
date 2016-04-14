package service

import (
	"io"
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/codes"
	"github.com/clawio/entities"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// Download streams a file to the client.
func (s *Service) Download(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	user := context.Get(r, userKey).(entities.User)
	reader, err := s.DataController.DownloadBLOB(user, path)
	if err != nil {
		s.handleDownloadError(err, w)
		return
	}
	if _, err := io.Copy(w, reader); err != nil {
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Error("error writing response body")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Service) handleDownloadError(err error, w http.ResponseWriter) {
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}
	server.Log.WithFields(logrus.Fields{
		"error": err,
	}).Error("error downloading blob")
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return
}
