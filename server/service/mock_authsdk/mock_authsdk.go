package mock_authsdk

import (
	"github.com/clawio/service-auth/server/spec"
	"github.com/stretchr/testify/mock"
)

type MockSDK struct {
	mock.Mock
}

func (m *MockSDK) Authenticate(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}
func (m *MockSDK) Verify(token string) (*spec.Identity, error) {
	args := m.Called(token)
	return args.Get(0).(*spec.Identity), args.Error(1)
}
