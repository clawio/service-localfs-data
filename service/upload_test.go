package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/clawio/codes"
	"github.com/stretchr/testify/require"
)

func (suite *TestSuite) TestUpload() {
	reader := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	suite.MockDataController.On("UploadBLOB").Once().Return(nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", reader)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), http.StatusCreated, w.Code)
}
func (suite *TestSuite) TestUpload_withNilBody() {
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 400, w.Code)
}
func (suite *TestSuite) TestUpload_withBodyTooBig() {
	reader := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	suite.MockDataController.On("UploadBLOB").Once().Return(errors.New("http: request body too large"))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", reader)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), http.StatusRequestEntityTooLarge, w.Code)
}
func (suite *TestSuite) TestUpload_withError() {
	reader := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	suite.MockDataController.On("UploadBLOB").Once().Return(errors.New("my test error"))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", reader)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}
func (suite *TestSuite) TestUpload_withCodeNotFound() {
	reader := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	suite.MockDataController.On("UploadBLOB").Once().Return(codes.NewErr(codes.NotFound, ""))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", reader)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), http.StatusNotFound, w.Code)
}
func (suite *TestSuite) TestUpload_withCodeBadChecksum() {
	reader := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(user, &codes.Response{}, nil)
	suite.MockDataController.On("UploadBLOB").Once().Return(codes.NewErr(codes.BadChecksum, ""))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/myblob", reader)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), http.StatusPreconditionFailed, w.Code)
}
func (suite *TestSuite) TestgetClientChecksum_header() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	r.Header.Set("checksum", "mychecksum")
	checksum := suite.Service.getClientChecksum(r)
	require.Equal(suite.T(), "mychecksum", checksum)
}
func (suite *TestSuite) TestgetClientChecksum_query() {
	r, err := http.NewRequest("GET", "/", nil)
	require.Nil(suite.T(), err)
	values := r.URL.Query()
	values.Set("checksum", "mychecksum")
	r.URL.RawQuery = values.Encode()
	checksum := suite.Service.getClientChecksum(r)
	require.Equal(suite.T(), "mychecksum", checksum)
}
