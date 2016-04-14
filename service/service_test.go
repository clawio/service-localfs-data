package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"errors"
	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
	"github.com/clawio/codes"
	emocks "github.com/clawio/entities/mocks"
	"github.com/clawio/sdk"
	"github.com/clawio/sdk/mocks"
	mock_datacontroller "github.com/clawio/service-localfs-data/datacontroller/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var user = &emocks.MockUser{Username: "test"}

type TestSuite struct {
	suite.Suite
	MockAuthService    *mocks.MockAuthService
	MockDataController *mock_datacontroller.MockDataController
	SDK                *sdk.SDK
	Service            *Service
	Server             *server.SimpleServer
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
	cfg := &Config{
		Server: &config.Server{},
		General: &GeneralConfig{
			RequestBodyMaxSize: 1024, // 1KiB
		},
		DataController: &DataControllerConfig{
			Type:          "simple",
			SimpleDataDir: "/tmp",
			SimpleTempDir: "/tmp",
		},
	}
	mockAuthService := &mocks.MockAuthService{}
	s := &sdk.SDK{}
	s.Auth = mockAuthService

	svc := &Service{}
	svc.SDK = s
	svc.Config = cfg

	mockDataController := &mock_datacontroller.MockDataController{}
	svc.DataController = mockDataController
	suite.MockDataController = mockDataController

	suite.Service = svc
	suite.MockAuthService = mockAuthService
	serv := server.NewSimpleServer(cfg.Server)
	serv.Register(suite.Service)
	suite.Server = serv
	// create homedir for user test
	err := os.MkdirAll("/tmp/t/test", 0755)
	require.Nil(suite.T(), err)

	// configure user mock
	user.On("GetUsername").Return("test")
}

func (suite *TestSuite) TeardownTest() {
	os.Remove("/tmp/t/test")
}

func (suite *TestSuite) TestNew() {
	cfg := &Config{
		Server: &config.Server{},
		General: &GeneralConfig{
			RequestBodyMaxSize: 1024, // 1KiB
		},
		DataController: &DataControllerConfig{
			Type:          "simple",
			SimpleDataDir: "/tmp",
			SimpleTempDir: "/tmp",
		},
	}
	svc, err := New(cfg)
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), svc)
}
func (suite *TestSuite) TestNew_withNilConfig() {
	_, err := New(nil)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestNew_withNilGeneralConfig() {
	cfg := &Config{
		Server:  nil,
		General: nil,
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestNew_withNilDataControllerConfig() {
	cfg := &Config{
		Server:         nil,
		General:        &GeneralConfig{},
		DataController: nil,
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestPrefix() {
	require.Equal(suite.T(), "/clawio/v1/data", suite.Service.Prefix())
}

func (suite *TestSuite) TestMetrics() {
	r, err := http.NewRequest("GET", "/clawio/v1/data/metrics", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 200, w.Code)
}
func (suite *TestSuite) TestgetTokenFromRequest_header() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("token", "mytoken")
	token := suite.Service.getTokenFromRequest(r)
	require.Equal(suite.T(), "mytoken", token)

}
func (suite *TestSuite) TestgetTokenFromRequest_query() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	values := r.URL.Query()
	values.Set("token", "mytoken")
	r.URL.RawQuery = values.Encode()
	token := suite.Service.getTokenFromRequest(r)
	require.Equal(suite.T(), "mytoken", token)
}
func (suite *TestSuite) TestAuthenticateHandlerFunc() {
	suite.MockAuthService.On("Verify", "mytoken").Once().Return(user, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", nil)
	r.Header.Set("token", "mytoken")
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.NotEqual(suite.T(), 401, w.Code)
}
func (suite *TestSuite) TestAuthenticateHandlerFunc_withBadToken() {
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, errors.New("test error"))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 401, w.Code)
}
