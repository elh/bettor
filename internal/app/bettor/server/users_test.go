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
	users := []*api.User{
		{
			Name:     "books/guild:other/users/asdf",
			Username: "rusty",
		},
	}
	testCases := []struct {
		desc      string
		book      string
		user      *api.User
		expectErr bool
	}{
		{
			desc: "basic case",
			book: entity.BookN("guild:A"),
			user: &api.User{
				Username: "rusty",
			},
		},
		{
			desc: "fails if username already exists in book",
			book: entity.BookN("guild:other"),
			user: &api.User{
				Username: "rusty",
			},
			expectErr: true,
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
			desc: "fails if book is invalid length",
			book: strings.Repeat("A", 37),
			user: &api.User{
				Username: "rusty",
			},
			expectErr: true,
		},
		{
			desc:      "fails if username is empty",
			book:      entity.BookN("guild:A"),
			user:      &api.User{},
			expectErr: true,
		},
		{
			desc:      "fails if username is too long",
			book:      entity.BookN("guild:A"),
			user:      &api.User{Username: strings.Repeat("A", 256)},
			expectErr: true,
		},
		{
			desc:      "fails if username is not alphanumeric",
			book:      entity.BookN("guild:A"),
			user:      &api.User{Username: "ᵣᵤₛₜᵧ"},
			expectErr: true,
		},
		{
			desc:      "fails if user is nil",
			book:      entity.BookN("guild:A"),
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
			s, err := server.New(server.WithRepo(&mem.Repo{Users: users}))
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
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	userHydrated := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "linus",
		Centipoints: 100,
	}
	unsettledBet := &api.Bet{
		Name:        entity.BetN("guild:1", "a"),
		User:        userHydrated.Name,
		Centipoints: 200,
	}
	unsettledBet2 := &api.Bet{
		Name:        entity.BetN("guild:1", "b"),
		User:        userHydrated.Name,
		Centipoints: 50,
	}
	hydratedUserHydrated := proto.Clone(userHydrated).(*api.User)
	hydratedUserHydrated.UnsettledCentipoints += unsettledBet.Centipoints
	hydratedUserHydrated.UnsettledCentipoints += unsettledBet2.Centipoints
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
			desc:     "hydrate unsettled points",
			user:     userHydrated.GetName(),
			expected: hydratedUserHydrated,
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
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user, userHydrated}, Bets: []*api.Bet{unsettledBet, unsettledBet2}}))
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
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	userHydrated := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "linus",
		Centipoints: 100,
	}
	unsettledBet := &api.Bet{
		Name:        entity.BetN("guild:1", "a"),
		User:        userHydrated.Name,
		Centipoints: 200,
	}
	unsettledBet2 := &api.Bet{
		Name:        entity.BetN("guild:1", "b"),
		User:        userHydrated.Name,
		Centipoints: 50,
	}
	hydratedUserHydrated := proto.Clone(userHydrated).(*api.User)
	hydratedUserHydrated.UnsettledCentipoints += unsettledBet.Centipoints
	hydratedUserHydrated.UnsettledCentipoints += unsettledBet2.Centipoints
	testCases := []struct {
		desc      string
		book      string
		username  string
		expected  *api.User
		expectErr bool
	}{
		{
			desc:     "basic case",
			book:     entity.BookN("guild:1"),
			username: "rusty",
			expected: user,
		},
		{
			desc:     "hydrate unsettled points",
			book:     entity.BookN("guild:1"),
			username: "linus",
			expected: hydratedUserHydrated,
		},
		{
			desc:      "fails if user does not exist",
			book:      entity.BookN("guild:1"),
			username:  "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			book:      entity.BookN("guild:1"),
			username:  "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user, userHydrated}, Bets: []*api.Bet{unsettledBet, unsettledBet2}}))
			require.Nil(t, err)
			out, err := s.GetUserByUsername(context.Background(), connect.NewRequest(&api.GetUserByUsernameRequest{Book: tC.book, Username: tC.username}))
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
	bookID := "Z"
	user1 := &api.User{
		Name:        entity.UserN(bookID, "a"),
		Username:    "rusty",
		Centipoints: 100,
	}
	user2 := &api.User{
		Name:        entity.UserN(bookID, "b"),
		Username:    "danny",
		Centipoints: 200,
	}
	// has an unsettled bet
	user3 := &api.User{
		Name:        entity.UserN(bookID, "c"),
		Username:    "linus",
		Centipoints: 300,
	}
	unsettledBet := &api.Bet{
		Name:        entity.BetN(bookID, "a"),
		User:        user3.Name,
		Centipoints: 200,
	}
	user3Hydrated := proto.Clone(user3).(*api.User)
	user3Hydrated.UnsettledCentipoints += unsettledBet.Centipoints
	testCases := []struct {
		desc          string
		req           *api.ListUsersRequest
		expected      []*api.User
		expectedCalls int
		expectErr     bool
	}{
		{
			desc:          "basic case",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID)},
			expected:      []*api.User{user1, user2, user3Hydrated},
			expectedCalls: 1,
		},
		{
			desc:          "searches within book",
			req:           &api.ListUsersRequest{Book: entity.BookN("other")},
			expected:      nil,
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID), PageSize: 1},
			expected:      []*api.User{user1, user2, user3Hydrated},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID), PageSize: 2},
			expected:      []*api.User{user1, user2, user3Hydrated},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID), PageSize: 3},
			expected:      []*api.User{user1, user2, user3Hydrated},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID), PageSize: 4},
			expected:      []*api.User{user1, user2, user3Hydrated},
			expectedCalls: 1,
		},
		{
			desc:          "list by user resource names",
			req:           &api.ListUsersRequest{Book: entity.BookN(bookID), Users: []string{user1.Name, user2.Name}},
			expected:      []*api.User{user1, user2},
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user1, user2, user3}, Bets: []*api.Bet{unsettledBet}}))
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
