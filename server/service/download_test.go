package service

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/clawio/service-auth/server/spec"
	"github.com/stretchr/testify/require"
)

func (suite *TestSuite) TestDownloadInvalidOrEmptyToken() {
	suite.MockAuthSDK.On("Verify", "").Once().Return(&spec.Identity{}, errors.New("test error"))
	r, err := http.NewRequest("GET", "/clawio/data/v1/download/testresource", nil)
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
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("GET", "/clawio/data/v1/download/thisfiledoesnotexists", nil)
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
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("GET", "/clawio/data/v1/download/thisfiledoesnotexists", nil)
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
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("GET", "/clawio/data/v1/download/samplefile", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 200, w.Code)
}

/*
func (suite *TestSuite) TestDownloadNilBody() {
	suite.MockAuthSDK.On("Verify", "").Once().Return(&spec.Identity{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", nil)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 400, w.Code)
}

func (suite *TestSuite) TestDownloadUnexistentTempDir() {
	suite.Service.Config.Storage.TempDir = "/tmp/this/not/exists"
	body := strings.NewReader("1")
	suite.MockAuthSDK.On("Verify", "").Once().Return(&spec.Identity{}, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestDownloadHomeDirNotCreated() {
	err := os.RemoveAll(path.Join(suite.Service.Config.Storage.DataDir, "t", "test"))
	require.Nil(suite.T(), err)
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestDownloadBadServerChecksumType() {
	suite.Service.Config.Storage.Checksum = "someinventedchekcsumtype"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 500, w.Code)
}

func (suite *TestSuite) TestDownloadInvalidClientChecksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	suite.Service.Config.Storage.VerifyClientChecksum = true
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource?checksum=abc", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 412, w.Code)
}

func (suite *TestSuite) TestDownload() {
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "testtoken").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	r.Header.Set("token", "testtoken")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
}

func (suite *TestSuite) TestDownloadMD5Checksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// md5 checksum of 1 is c4ca4238a0b923820dcc509a6f75849b
	require.Equal(suite.T(), "md5:c4ca4238a0b923820dcc509a6f75849b", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestDownloadAdler32Checksum() {
	suite.Service.Config.Storage.Checksum = "adler32"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// adler32 checksum of 1 is 00320032
	require.Equal(suite.T(), "adler32:00320032", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestDownloadSha1Checksum() {
	suite.Service.Config.Storage.Checksum = "sha1"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// sha1 checksum of 1 is 356a192b7913b04c54574d18c28d46e6395428ab
	require.Equal(suite.T(), "sha1:356a192b7913b04c54574d18c28d46e6395428ab", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestDownloadSha256Checksum() {
	suite.Service.Config.Storage.Checksum = "sha256"
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
	// sha1 checksum of 1 is 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b
	require.Equal(suite.T(), "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b", w.Header().Get("checksum"))
}

func (suite *TestSuite) TestDownloadClientChecksum() {
	suite.Service.Config.Storage.Checksum = "md5"
	suite.Service.Config.Storage.VerifyClientChecksum = true
	body := strings.NewReader("1")
	testIdentity := &spec.Identity{
		Username:    "test",
		Email:       "test@test.com",
		DisplayName: "Test",
	}
	suite.MockAuthSDK.On("Verify", "").Once().Return(testIdentity, nil)
	r, err := http.NewRequest("PUT", "/clawio/data/v1/upload/testresource", body)
	require.Nil(suite.T(), err)
	r.Header.Set("checksum", "md5:c4ca4238a0b923820dcc509a6f75849b")
	w := httptest.NewRecorder()
	suite.Server.ServeHTTP(w, r)
	require.Equal(suite.T(), 201, w.Code)
}*/
