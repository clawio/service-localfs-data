package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NYTimes/gizmo/config"
	"github.com/NYTimes/gizmo/server"
	"github.com/clawio/service-localfs-data/server/service/mock_authsdk"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	MockAuthSDK *mock_authsdk.MockSDK
	Service     *Service
	Server      *server.SimpleServer
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (suite *TestSuite) SetupTest() {
	mockAuthSDK := &mock_authsdk.MockSDK{}
	cfg := &Config{
		Server: &config.Server{},
		Storage: &Storage{
			TempDir:            "/tmp",
			DataDir:            "/tmp",
			RequestBodyMaxSize: 1024, // 1KB
		},
	}
	svc := &Service{}
	svc.AuthNSDK = mockAuthSDK
	svc.Config = cfg
	suite.Service = svc
	suite.MockAuthSDK = mockAuthSDK
	serv := server.NewSimpleServer(cfg.Server)
	serv.Register(suite.Service)
	suite.Server = serv
	// create homedir for user test
	err := os.MkdirAll("/tmp/t/test", 0755)
	require.Nil(suite.T(), err)
}

func (suite *TestSuite) TeardownTest() {
	os.Remove("/tmp/t/test")
}

func (suite *TestSuite) TestNew() {
	storageCfg := &Storage{
		DataDir: "/tmp/datadir",
		TempDir: "/tmp",
	}
	cfg := &Config{
		Server:  nil,
		Storage: storageCfg,
	}
	svc, err := New(cfg)
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), svc)
}
func (suite *TestSuite) TestNewNilConfig() {
	_, err := New(nil)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestNewNilStorageConfig() {
	cfg := &Config{
		Server:  nil,
		Storage: nil,
	}
	_, err := New(cfg)
	require.NotNil(suite.T(), err)
}

func (suite *TestSuite) TestPrefix() {
	require.Equal(suite.T(), "/clawio/data/v1", suite.Service.Prefix())
}

func (suite *TestSuite) TestMetrics() {
	r, err := http.NewRequest("GET", "/clawio/data/v1/metrics", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 200, w.Code)
}

func (suite *TestSuite) TestSecureJoin() {
	paths := []struct {
		given    []string
		expected string
	}{
		{
			[]string{"relativePath/t/test"},
			"relativePath/t/test",
		},
		{
			[]string{"../../relativePath/t/test"},
			"../../relativePath/t/test",
		},
		{
			[]string{"../../relativePath/t/test", "../../../../"},
			"../../relativePath/t/test",
		},
		{
			[]string{"/abspath/t/test"},
			"/abspath/t/test",
		},
		{
			[]string{"/abspath/t/test", "../../.."},
			"/abspath/t/test",
		},
	}

	for _, v := range paths {
		require.Equal(suite.T(), v.expected, secureJoin(v.given...))
	}
}
