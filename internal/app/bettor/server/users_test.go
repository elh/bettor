package server_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/entity"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		desc      string
		book      string
		user      *api.User
		expectErr bool
	}{
		{
			desc: "basic case",
			book: "guild:A",
			user: &api.User{
				Username: "rusty",
			},
		},
		{
			desc: "fails if book is invalid (has a slash)",
			book: "guild/A",
			user: &api.User{
				Username: "rusty",
			},
			expectErr: true,
		},
		{
			desc:      "fails if username is empty",
			book:      "guild:A",
			user:      &api.User{},
			expectErr: true,
		},
		{
			desc:      "fails if username is too long",
			book:      "guild:A",
			user:      &api.User{Username: strings.Repeat("A", 256)},
			expectErr: true,
		},
		{
			desc:      "fails if username is not alphanumeric",
			book:      "guild:A",
			user:      &api.User{Username: "ᵣᵤₛₜᵧ"},
			expectErr: true,
		},
		{
			desc:      "fails if user is nil",
			book:      "guild:A",
			user:      nil,
			expectErr: true,
		},
		{
			desc: "fails is book is not set",
			book: "",
			user: &api.User{
				Username: "rusty",
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{}))
			require.Nil(t, err)
			out, err := s.CreateUser(context.Background(), connect.NewRequest(&api.CreateUserRequest{Book: tC.book, User: tC.user}))
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
		Name:        entity.UserN("A", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc      string
		user      string
		expected  *api.User
		expectErr bool
	}{
		{
			desc:     "basic case",
			user:     user.GetName(),
			expected: user,
		},
		{
			desc:      "fails if user does not exist",
			user:      "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			user:      "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user}}))
			require.Nil(t, err)
			out, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: tC.user}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetUser())
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	user := &api.User{
		Name:        entity.UserN("A", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc      string
		username  string
		expected  *api.User
		expectErr bool
	}{
		{
			desc:     "basic case",
			username: "rusty",
			expected: user,
		},
		{
			desc:      "fails if user does not exist",
			username:  "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			username:  "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user}}))
			require.Nil(t, err)
			out, err := s.GetUserByUsername(context.Background(), connect.NewRequest(&api.GetUserByUsernameRequest{Username: tC.username}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetUser())
		})
	}
}

func TestListUsers(t *testing.T) {
	// tests pagination until all users are returned
	// alphabetically ordered ids
	user1 := &api.User{
		Name:        entity.UserN("Z", "a"),
		Username:    "rusty",
		Centipoints: 100,
	}
	user2 := &api.User{
		Name:        entity.UserN("Z", "b"),
		Username:    "danny",
		Centipoints: 200,
	}
	user3 := &api.User{
		Name:        entity.UserN("Z", "c"),
		Username:    "linus",
		Centipoints: 300,
	}
	testCases := []struct {
		desc          string
		req           *api.ListUsersRequest
		expected      []*api.User
		expectedCalls int
		expectErr     bool
	}{
		{
			desc:          "basic case",
			req:           &api.ListUsersRequest{},
			expected:      []*api.User{user1, user2, user3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListUsersRequest{PageSize: 1},
			expected:      []*api.User{user1, user2, user3},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListUsersRequest{PageSize: 2},
			expected:      []*api.User{user1, user2, user3},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListUsersRequest{PageSize: 3},
			expected:      []*api.User{user1, user2, user3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListUsersRequest{PageSize: 4},
			expected:      []*api.User{user1, user2, user3},
			expectedCalls: 1,
		},
		{
			desc:          "list by user resource names",
			req:           &api.ListUsersRequest{Users: []string{user1.Name, user2.Name}},
			expected:      []*api.User{user1, user2},
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user1, user2, user3}}))
			require.Nil(t, err)
			var all []*api.User
			var calls int
			var pageToken string
			for {
				req := proto.Clone(tC.req).(*api.ListUsersRequest)
				req.PageToken = pageToken
				out, err := s.ListUsers(context.Background(), connect.NewRequest(req))
				if tC.expectErr {
					require.NotNil(t, err)
					return
				}
				calls++
				require.Nil(t, err)
				require.NotNil(t, out)
				all = append(all, out.Msg.GetUsers()...)
				if out.Msg.GetNextPageToken() == "" {
					break
				}
				pageToken = out.Msg.GetNextPageToken()
			}
			assert.Equal(t, tC.expected, all)
			assert.Equal(t, tC.expectedCalls, calls)
		})
	}
}
