package mockclient

import (
	"fenix/src/test_utils"
	"testing"
)

type Credentials struct {
	Username string
	Password string
}

func (m *MockClient) RegisterClient(t *testing.T, srv *test_utils.ServerFields, auth Credentials) *test_utils.ClientFields {
	t.Helper()

	srv.Addr.Path = "/register"
	cli := test_utils.Connect(auth.Username, auth.Password, srv.Addr)

	return cli
}

func (m *MockClient) LoginClient(t *testing.T, srv *test_utils.ServerFields, auth Credentials) *test_utils.ClientFields {
	t.Helper()

	srv.Addr.Path = "/login"
	cli := test_utils.Connect(auth.Username, auth.Password, srv.Addr)

	return cli
}
