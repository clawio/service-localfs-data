package service

import (
	"net/http"

	"github.com/NYTimes/gizmo/config"
	"github.com/clawio/service-auth/sdk"
	"github.com/prometheus/client_golang/prometheus"
)

type (

	// Service will implement server.Service and
	// handle all requests to the server.
	Service struct {
		Config   *Config
		AuthNSDK *sdk.SDK
	}

	// Config is a struct to contain all the needed
	// configuration for our Service
	Config struct {
		Server  *config.Server
		Storage *Storage
	}

	// Storage is a struct that contains all
	// Storage configuration parameters.
	Storage struct {
		DataDir       string
		TempDir       string
		Checksum      string
		PropagatorURL string
		AuthNURL      string
	}
)

// New will instantiate and return
// a new Service that implements server.Service.
func New(cfg *Config) *Service {
	authNSDK := sdk.New(cfg.Storage.AuthNURL, nil)
	return &Service{Config: cfg, AuthNSDK: authNSDK}
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *Service) Prefix() string {
	return "/clawio/data/v1"
}

// Middleware provides an http.Handler hook wrapped around all requests.
// In this implementation, we authenticate the request.
func (s *Service) Middleware(h http.Handler) http.Handler {
	return h
}

// Endpoints is a listing of all endpoints available in the Service.
func (s *Service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/metrics": map[string]http.HandlerFunc{
			"GET": func(w http.ResponseWriter, r *http.Request) {
				prometheus.Handler().ServeHTTP(w, r)
			},
		},
		"/upload/{path:.*}": map[string]http.HandlerFunc{
			"PUT": prometheus.InstrumentHandlerFunc("/upload", s.AuthenticateHandlerFunc(s.Upload)),
		},
		"/download/{path:.*}": map[string]http.HandlerFunc{
			"GET": prometheus.InstrumentHandlerFunc("/download", s.AuthenticateHandlerFunc(s.Download)),
		},
	}
}

func (s *Service) getTokenFromRequest(r *http.Request) string {
	if t := r.Header.Get("token"); t != "" {
		return t
	}
	return r.URL.Query().Get("token")
}

func (s *Service) AuthenticateHandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := s.getTokenFromRequest(r)
		_, err := s.AuthNSDK.Verify(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}
