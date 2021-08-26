package mocks

import "github.com/sy-software/minerva-olive/internal/core/ports"

type MockSecrets struct {
	Values map[string]string
}

func (mock *MockSecrets) Get(name string) (string, error) {
	val, ok := mock.Values[name]

	if !ok {
		return "", ports.ErrSecretNoExists
	}

	return val, nil
}
