package mock

import (
	"io"

	"github.com/clawio/service-auth/server/spec"
	"github.com/stretchr/testify/mock"
)

type MockDataController struct {
	mock.Mock
}

func (m *MockDataController) UploadBLOB(user *spec.Identity, pathSpec string, r io.Reader, clientChecksum string) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockDataController) DownloadBLOB(user *spec.Identity, pathSpec string) (io.Reader, error) {
	args := m.Called()
	return args.Get(0).(io.Reader), args.Error(1)
}
