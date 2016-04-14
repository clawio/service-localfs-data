package mock

import (
	"io"

	"github.com/clawio/entities"
	"github.com/stretchr/testify/mock"
)

type MockDataController struct {
	mock.Mock
}

func (m *MockDataController) UploadBLOB(user entities.User, pathSpec string, r io.Reader, clientChecksum string) error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockDataController) DownloadBLOB(user entities.User, pathSpec string) (io.Reader, error) {
	args := m.Called()
	return args.Get(0).(io.Reader), args.Error(1)
}
