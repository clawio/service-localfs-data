package service

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"

	"github.com/clawio/codes"
	"github.com/clawio/service-auth/server/spec"
	"github.com/stretchr/testify/require"
)

func (suite *TestSuite) TestUploadInvalidOrEmptyToken() {
	suite.MockAuthService.On("Verify", "").Once().Return(&spec.Identity{}, &codes.Response{}, errors.New("test error"))
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 401, w.Code)
}

func (suite *TestSuite) TestUploadNilBody() {
	suite.MockAuthService.On("Verify", "").Once().Return(&spec.Identity{}, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestUploadUnexistentTempDir() {
	suite.Service.Config.Storage.TempDir = "/tmp/this/not/exists"
	body := strings.NewReader("1")
	suite.MockAuthService.On("Verify", "").Once().Return(&spec.Identity{}, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestUploadHomeDirNotCreated() {
	err := os.RemoveAll(path.Join(suite.Service.Config.Storage.DataDir, "t", "test"))
	require.Nil(suite.T(), err)
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestUploadBadServerChecksumType() {
	suite.Service.Config.Storage.Checksum = "someinventedchekcsumtype"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestUploadInvalidClientChecksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	suite.Service.Config.Storage.VerifyClientChecksum = true
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource?checksum=abc", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 412, w.Code)
}

func (suite *TestSuite) TestUploadFileTooBig() {
	suite.Service.Config.Storage.RequestBodyMaxSize = 2 // 2 bytes
	body := strings.NewReader("12345678910toobigfile")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "testtoken").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	r.Header.Set("token", "testtoken")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 413, w.Code)
}

func (suite *TestSuite) TestUpload() {
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "testtoken").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	r.Header.Set("token", "testtoken")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
}

func (suite *TestSuite) TestUploadMD5Checksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// md5 checksum of 1 is c4ca4238a0b923820dcc509a6f75849b
	require.Equal(suite.T(), "md5:c4ca4238a0b923820dcc509a6f75849b", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestUploadAdler32Checksum() {
	suite.Service.Config.Storage.Checksum = "adler32"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// adler32 checksum of 1 is 00320032
	require.Equal(suite.T(), "adler32:00320032", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestUploadSha1Checksum() {
	suite.Service.Config.Storage.Checksum = "sha1"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// sha1 checksum of 1 is 356a192b7913b04c54574d18c28d46e6395428ab
	require.Equal(suite.T(), "sha1:356a192b7913b04c54574d18c28d46e6395428ab", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestUploadSha256Checksum() {
	suite.Service.Config.Storage.Checksum = "sha256"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// sha1 checksum of 1 is 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b
	require.Equal(suite.T(), "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestUploadClientChecksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	suite.Service.Config.Storage.VerifyClientChecksum = true
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthService.On("Verify", "").Once().Return(testIdentity, &codes.Response{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/v1/data/upload/testresource", body)
	require.Nil(suite.T(), err)
	r.Header.Set("checksum", "md5:c4ca4238a0b923820dcc509a6f75849b")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
}
