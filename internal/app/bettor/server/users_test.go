package server_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *api.User
		expectErr bool
	}{
		{
			desc: "dummy test",
			user: &api.User{},
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New()
			_, err := s.CreateUser(context.Background(), connect.NewRequest(&api.CreateUserRequest{User: tC.user}))
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
			desc:      "unimplemented",
			userID:    "rusty",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New()
			_, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: tC.userID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}
