package service

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Upload saves a file to disk.
func (s *Service) Upload(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	w.Write([]byte(path))
}
