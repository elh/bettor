package server_test

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMarket(t *testing.T) {
	user := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc      string
		market    *api.Market
		expectErr bool
	}{
		{
			desc: "basic case",
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.Id,
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: []*api.Outcome{
							{Title: "Yes"},
							{Title: "No"},
						},
					},
				},
			},
		},
		{
			desc: "fails if title not set",
			market: &api.Market{
				Creator: user.Id,
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: []*api.Outcome{
							{Title: "Yes"},
							{Title: "No"},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			desc: "fails if creator is not an existing user",
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: "other",
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: []*api.Outcome{
							{Title: "Yes"},
							{Title: "No"},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			desc: "fails if type not implemented",
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.Id,
			},
			expectErr: true,
		},
		{
			desc: "fails if pool has less than 2 outcomes",
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.Id,
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: []*api.Outcome{
							{Title: "Yes"},
						},
					},
				},
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{user}})
			out, err := s.CreateMarket(context.Background(), connect.NewRequest(&api.CreateMarketRequest{Market: tC.market}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)

			assert.NotEmpty(t, out)
		})
	}
}
