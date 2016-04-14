package datacontroller

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/clawio/entities/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var user = &mocks.MockUser{Username: "test"}

type TestSuite struct {
	suite.Suite
	dataController       DataController
	simpleDataController *simpleDataController
}

func Test(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
func (suite *TestSuite) SetupTest() {
	opts := &SimpleDataControllerOptions{
		DataDir: "/tmp",
		TempDir: "/tmp",
	}
	dataController := NewSimpleDataController(opts)
	// create homedir for user test
	err := os.MkdirAll("/tmp/t/test", 0755)
	require.Nil(suite.T(), err)
	suite.dataController = dataController
	suite.simpleDataController = suite.dataController.(*simpleDataController)

	// configure user mock
	user.On("GetUsername").Return("test")
}
func (suite *TestSuite) TeardownTest() {
	os.RemoveAll("/tmp/t")
}
func (suite *TestSuite) TestNewSimpleDataController() {
	opts := &SimpleDataControllerOptions{
		DataDir: "/tmp",
		TempDir: "/tmp",
	}
	require.IsType(suite.T(), &simpleDataController{}, NewSimpleDataController(opts))
}
func (suite *TestSuite) TestNewSimpleDataController_withNilOptions() {
	require.IsType(suite.T(), &simpleDataController{}, NewSimpleDataController(nil))
}
func (suite *TestSuite) TestUpload() {
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "")
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withBadTempDir() {
	suite.simpleDataController.tempDir = "/this/does/not/exist"
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withChecksum() {
	suite.simpleDataController.checksum = "md5"
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "")
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withWrongChecksum() {
	suite.simpleDataController.checksum = "xyz"
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withClientChecksum() {
	suite.simpleDataController.checksum = "md5"
	suite.simpleDataController.verifyClientChecksum = true
	reader := strings.NewReader("1")
	// md5 checksum of 1 is c4ca4238a0b923820dcc509a6f75849b
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "md5:c4ca4238a0b923820dcc509a6f75849b")
	require.Nil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withWrongClientChecksum() {
	suite.simpleDataController.checksum = "md5"
	suite.simpleDataController.verifyClientChecksum = true
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "md5:")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestUpload_withBadDataDir() {
	suite.simpleDataController.dataDir = "/this/does/not/exist"
	reader := strings.NewReader("1")
	err := suite.dataController.UploadBLOB(user, "myblob", reader, "")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestDownload() {
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	reader, err := suite.dataController.DownloadBLOB(user, "myblob")
	require.Nil(suite.T(), err)
	data, err := ioutil.ReadAll(reader)
	require.Nil(suite.T(), err)
	require.Equal(suite.T(), "1", string(data))
}
func (suite *TestSuite) TestDownload_withBadDataDir() {
	suite.simpleDataController.dataDir = "/this/does/not/exist"
	_, err := suite.dataController.DownloadBLOB(user, "myblob")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestcomputeChecksum_withBadChecksum() {
	suite.simpleDataController.checksum = "xyz"
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	_, err = suite.simpleDataController.computeChecksum(p)
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestcomputeChecksum_withNoFile() {
	suite.simpleDataController.checksum = "md5"
	_, err := suite.simpleDataController.computeChecksum("/this/does/not/exist/myblob")
	require.NotNil(suite.T(), err)
}
func (suite *TestSuite) TestcomputeChecksum_md5() {
	suite.simpleDataController.checksum = "md5"
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	checksum, err := suite.simpleDataController.computeChecksum(p)
	require.Nil(suite.T(), err)

	// md5 checksum of "1" is c4ca4238a0b923820dcc509a6f75849b
	require.Equal(suite.T(), "md5:c4ca4238a0b923820dcc509a6f75849b", checksum)
}
func (suite *TestSuite) TestcomputeChecksum_adler32() {
	suite.simpleDataController.checksum = "adler32"
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	checksum, err := suite.simpleDataController.computeChecksum(p)
	require.Nil(suite.T(), err)

	// adler32 checksum of "1" is 00320032
	require.Equal(suite.T(), "adler32:00320032", checksum)
}
func (suite *TestSuite) TestcomputeChecksum_sha1() {
	suite.simpleDataController.checksum = "sha1"
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	checksum, err := suite.simpleDataController.computeChecksum(p)
	require.Nil(suite.T(), err)

	// sha1 checksum of "1" is 356a192b7913b04c54574d18c28d46e6395428ab
	require.Equal(suite.T(), "sha1:356a192b7913b04c54574d18c28d46e6395428ab", checksum)
}
func (suite *TestSuite) TestcomputeChecksum_sha256() {
	suite.simpleDataController.checksum = "sha256"
	p := path.Join(suite.simpleDataController.tempDir, "t", "test", "myblob")
	err := ioutil.WriteFile(p, []byte("1"), 0644)
	require.Nil(suite.T(), err)
	checksum, err := suite.simpleDataController.computeChecksum(p)
	require.Nil(suite.T(), err)

	// sha256 checksum of "1" is 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b
	require.Equal(suite.T(), "sha256:6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b", checksum)
}

func (suite *TestSuite) TestsecureJoin() {
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
