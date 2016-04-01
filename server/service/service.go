package service

import (
	"errors"
	"net/http"
	"path"

	"github.com/NYTimes/gizmo/config"
	"github.com/clawio/sdk"
	"github.com/clawio/service-auth/server/spec"
	"github.com/gorilla/context"
	"github.com/prometheus/client_golang/prometheus"
)

// identityKey can be used to store/retrieve a user ID in a request context.
const identityKey idKey = 0

type (
	// idKey is a type to use as a key for storing data in the request context.
	idKey int

	// Service will implement server.Service and
	// handle all requests to the server.
	Service struct {
		Config *Config
		SDK    *sdk.SDK
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
		DataDir              string
		TempDir              string
		Checksum             string
		PropagatorURL        string
		AuthNURL             string
		VerifyClientChecksum bool
		RequestBodyMaxSize   int64
	}
)

// New will instantiate and return
// a new Service that implements server.Service.
func New(cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}
	if cfg.Storage == nil {
		return nil, errors.New("config.storage cannot be nil")
	}
	urls := &sdk.ServiceEndpoints{}
	urls.AuthServiceBaseURL = cfg.Storage.AuthNURL
	s := sdk.New(urls, nil)
	return &Service{Config: cfg, SDK: s}, nil
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
		identity, _, err := s.SDK.Auth.Verify(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		context.Set(r, identityKey, identity)
		handler(w, r)
	}
}

func (s *Service) getStoragePath(identity *spec.Identity, path string) string {
	homeDir := secureJoin("/", string(identity.Username[0]), identity.Username)
	userPath := secureJoin(homeDir, path)
	return secureJoin(s.Config.Storage.DataDir, userPath)
}

// secureJoin avoids path traversal attacks when joinning paths.
func secureJoin(args ...string) string {
	if len(args) > 1 {
		s := []string{"/"}
		s = append(s, args[1:]...)
		jailedPath := path.Join(s...)
		return path.Join(args[0], jailedPath)
	}
	return path.Join(args...)
}
