package service

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Download downloads a file from EOS
func (s *Service) Download(w http.ResponseWriter, r *http.Request) {
	path := mux.Vars(r)["path"]
	w.Write([]byte(path))
}
