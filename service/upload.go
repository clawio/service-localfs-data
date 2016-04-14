package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/server"
	"github.com/Sirupsen/logrus"
	"github.com/clawio/codes"
	"github.com/clawio/entities"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	path := mux.Vars(r)["path"]
	user := context.Get(r, userKey).(entities.User)
	clientChecksum := s.getClientChecksum(r)
	readCloser := http.MaxBytesReader(w, r.Body, s.Config.General.RequestBodyMaxSize)
	if err := s.DataController.UploadBLOB(user, path, readCloser, clientChecksum); err != nil {
		s.handleUploadError(err, w)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *Service) handleUploadError(err error, w http.ResponseWriter) {
	if err.Error() == "http: request body too large" {
		server.Log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("request body max size exceed")
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	if codeErr, ok := err.(*codes.Err); ok {
		if codeErr.Code == codes.NotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		if codeErr.Code == codes.BadChecksum {
			server.Log.WithFields(logrus.Fields{
				"error": err,
			}).Warn("blob corruption")
			http.Error(w, http.StatusText(http.StatusPreconditionFailed), http.StatusPreconditionFailed)
			return
		}
	}
	server.Log.WithFields(logrus.Fields{
		"error": err,
	}).Error("error saving blob")
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	return
}

func (s *Service) getClientChecksum(r *http.Request) string {
	if t := r.Header.Get("checksum"); t != "" {
		return t
	}
	return r.URL.Query().Get("checksum")
}
