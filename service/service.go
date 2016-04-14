package service

import (
	"errors"
	"net/http"

	"github.com/NYTimes/gizmo/config"
	"github.com/clawio/sdk"
	"github.com/clawio/service-localfs-data/datacontroller"
	"github.com/gorilla/context"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// userKey can be used to store/retrieve a user ID in a request context.
	userKey contextKey = iota
)

type (
	// contextKey is a type to use as a key for storing data in the request context.
	contextKey int

	// Service implements server.Service and
	// handle all requests to the server.
	Service struct {
		Config         *Config
		SDK            *sdk.SDK
		DataController datacontroller.DataController
	}

	// Config is a struct that holds the
	// configuration for Service
	Config struct {
		Server         *config.Server
		General        *GeneralConfig
		DataController *DataControllerConfig
	}

	// GeneralConfig contains configuration parameters
	// for general parts of the service.
	GeneralConfig struct {
		AuthenticationServiceBaseURL string
		RequestBodyMaxSize           int64
	}

	// DataControllerConfig is a struct that holds
	// configuration parameters for a data controller.
	DataControllerConfig struct {
		Type                       string
		SimpleDataDir              string
		SimpleTempDir              string
		SimpleChecksum             string
		SimpleVerifyClientChecksum bool
	}
)

// New will instantiate and return
// a new Service that implements server.Service.
func New(cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.General == nil {
		return nil, errors.New("config.General is nil")
	}
	if cfg.DataController == nil {
		return nil, errors.New("config.DataController is  nil")
	}

	urls := &sdk.ServiceEndpoints{}
	urls.AuthServiceBaseURL = cfg.General.AuthenticationServiceBaseURL
	s := sdk.New(urls, nil)

	dataController := getDataController(cfg.DataController)
	return &Service{Config: cfg, SDK: s, DataController: dataController}, nil
}

func getDataController(cfg *DataControllerConfig) datacontroller.DataController {
	opts := &datacontroller.SimpleDataControllerOptions{
		DataDir:              cfg.SimpleDataDir,
		TempDir:              cfg.SimpleTempDir,
		Checksum:             cfg.SimpleChecksum,
		VerifyClientChecksum: cfg.SimpleVerifyClientChecksum,
	}
	return datacontroller.NewSimpleDataController(opts)
}

// Prefix returns the string prefix used for all endpoints within
// this service.
func (s *Service) Prefix() string {
	return "/clawio/v1/data"
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
		user, _, err := s.SDK.Auth.Verify(token)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		context.Set(r, userKey, user)
		handler(w, r)
	}
}
