package server_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		desc      string
		user      *api.User
		expectErr bool
	}{
		{
			desc: "basic case",
			user: &api.User{
				Username: "rusty",
			},
		},
		{
			desc:      "fails if username is empty",
			user:      &api.User{},
			expectErr: true,
		},
		{
			desc:      "fails if username is too long",
			user:      &api.User{Username: strings.Repeat("A", 256)},
			expectErr: true,
		},
		{
			desc:      "fails if username is not alphanumeric",
			user:      &api.User{Username: "ᵣᵤₛₜᵧ"},
			expectErr: true,
		},
		{
			desc:      "fails if user is nil",
			user:      nil,
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{})
			out, err := s.CreateUser(context.Background(), connect.NewRequest(&api.CreateUserRequest{User: tC.user}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)

			assert.NotEmpty(t, out)
		})
	}
}

func TestGetUser(t *testing.T) {
	user := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc      string
		userID    string
		expected  *api.User
		expectErr bool
	}{
		{
			desc:     "basic case",
			userID:   user.Id,
			expected: user,
		},
		{
			desc:      "fails if user does not exist",
			userID:    "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			userID:    "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{user}})
			out, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: tC.userID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetUser())
		})
	}
}
