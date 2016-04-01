package service

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/clawio/codes"
	"github.com/clawio/service-auth/server/spec"
	"github.com/stretchr/testify/require"
)

func (suite *TestSuite) TestDownloadInvalidOrEmptyToken() {
	suite.MockAuthService.On("Verify", "").Once().Return(&spec.Identity{}, &codes.Response{}, errors.New("test error"))
	r, err := http.NewRequest("GET", "/clawio/v1/data/download/testresource", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 401, w.Code)
}

func (suite *TestSuite) TestDownloadFileNotFound() {
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("GET", "/clawio/v1/data/download/thisfiledoesnotexists", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 404, w.Code)
}

func (suite *TestSuite) TestDownloadOpenError() {
	err := os.MkdirAll("/tmp/testdir", 0000)
	defer os.RemoveAll("/tmp/testdir")
	require.Nil(suite.T(), err)
	suite.Service.Config.Storage.DataDir = "/tmp/testdir" // we should not have privileges to access it
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("GET", "/clawio/v1/data/download/thisfiledoesnotexists", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestDownload() {
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	fn := path.Join(suite.Service.Config.Storage.DataDir, "t/test/samplefile")
	err := ioutil.WriteFile(fn, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("GET", "/clawio/v1/data/download/samplefile", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 200, w.Code)
}
