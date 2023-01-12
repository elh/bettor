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

func TestGetMarket(t *testing.T) {
	market := &api.Market{
		Id: uuid.NewString(),
	}
	testCases := []struct {
		desc      string
		marketID  string
		expected  *api.Market
		expectErr bool
	}{
		{
			desc:     "basic case",
			marketID: market.Id,
			expected: market,
		},
		{
			desc:      "fails if market does not exist",
			marketID:  "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			marketID:  "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Markets: []*api.Market{market}})
			out, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: tC.marketID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetMarket())
		})
	}
}

func TestLockMarket(t *testing.T) {
	testCases := []struct {
		desc      string
		marketID  string
		expected  *api.Market
		expectErr bool
	}{
		{
			desc:      "unimplemented",
			marketID:  "todo",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{})
			out, err := s.LockMarket(context.Background(), connect.NewRequest(&api.LockMarketRequest{MarketId: tC.marketID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetMarket())
		})
	}
}

func TestSettleMarket(t *testing.T) {
	testCases := []struct {
		desc      string
		marketID  string
		expected  *api.Market
		expectErr bool
	}{
		{
			desc:      "unimplemented",
			marketID:  "todo",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{})
			out, err := s.SettleMarket(context.Background(), connect.NewRequest(&api.SettleMarketRequest{MarketId: tC.marketID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetMarket())
		})
	}
}

func TestCreateBet(t *testing.T) {
	user := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 1000,
	}
	poolMarket := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user.Id,
		Status:  api.Market_STATUS_ACTIVE,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Id: uuid.NewString(), Title: "Yes"},
					{Id: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	lockedPoolMarket := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user.Id,
		Status:  api.Market_STATUS_BETS_LOCKED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Id: uuid.NewString(), Title: "Yes"},
					{Id: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	settledPoolMarket := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user.Id,
		Status:  api.Market_STATUS_SETTLED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Id: uuid.NewString(), Title: "Yes"},
					{Id: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc                  string
		bet                   *api.Bet
		expectUserCentipoints uint64
		expectErr             bool
	}{
		// pool bets
		{
			desc: "basic case - pool bet",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			},
			expectUserCentipoints: 900,
		},
		{
			desc: "fails if user does not exist",
			bet: &api.Bet{
				UserId:      "other",
				MarketId:    poolMarket.Id,
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			},
			expectErr: true,
		},
		{
			desc: "fails if market does not exist",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    "other",
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			},
			expectErr: true,
		},
		{
			desc: "fails if type not provided",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 100,
			},
			expectErr: true,
		},
		{
			desc: "fails if outcome does not exist",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: "other"},
			},
			expectErr: true,
		},
		{
			desc: "fails if creating a bet on a locked market",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    lockedPoolMarket.Id,
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: lockedPoolMarket.GetPool().Outcomes[0].Id},
			},
			expectErr: true,
		},
		{
			desc: "fails if creating a bet on a settled market",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    settledPoolMarket.Id,
				Centipoints: 100,
				Type:        &api.Bet_OutcomeId{OutcomeId: settledPoolMarket.GetPool().Outcomes[0].Id},
			},
			expectErr: true,
		},
		{
			desc: "fails if betting more points than user has",
			bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 2000,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket, lockedPoolMarket, settledPoolMarket}})
			out, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{Bet: tC.bet}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)

			assert.NotEmpty(t, out)

			u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: tC.bet.GetUserId()}))
			require.Nil(t, err)
			assert.Equal(t, tC.expectUserCentipoints, u.Msg.User.GetCentipoints())

			m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: tC.bet.GetMarketId()}))
			require.Nil(t, err)
			assert.Equal(t, tC.bet.GetCentipoints(), m.Msg.Market.GetPool().GetOutcomes()[0].GetCentipoints())
		})
	}
}
