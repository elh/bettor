package server_test

import (
	"context"
	"testing"

	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/server"
	"gopkg.in/stretchr/testify.v1/require"
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *api.User
		expectErr bool
	}{
		{
			desc:      "currently unimplemented",
			user:      &api.User{},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s := &server.Server{}
			_, err := s.CreateUser(context.Background(), &api.CreateUserRequest{User: tC.user})
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestGetUser(t *testing.T) {
	testCases := []struct {
		desc      string
		userID    string
		expectErr bool
	}{
		{
			desc:      "currently unimplemented",
			userID:    "a",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s := &server.Server{}
			_, err := s.GetUser(context.Background(), &api.GetUserRequest{UserId: tC.userID})
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}
