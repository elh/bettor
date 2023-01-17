package server_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCreateMarket(t *testing.T) {
	var maxOpenMarkets []*api.Market
	for i := 0; i < server.MaxNumberOfOpenMarkets; i++ {
		maxOpenMarkets = append(maxOpenMarkets, &api.Market{
			Id:     uuid.NewString(),
			Status: api.Market_STATUS_OPEN,
		})
	}
	user := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc            string
		existingMarkets []*api.Market
		market          *api.Market
		expectErr       bool
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
		{
			desc:            "basic case",
			existingMarkets: maxOpenMarkets,
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
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: tC.existingMarkets}, log.NewNopLogger())
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
			s := server.New(&mem.Repo{Markets: []*api.Market{market}}, log.NewNopLogger())
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

func TestListMarkets(t *testing.T) {
	// tests pagination until all markets are returned
	// alphabetically ordered ids
	market1 := &api.Market{
		Id:     "a",
		Status: api.Market_STATUS_OPEN,
	}
	market2 := &api.Market{
		Id:     "b",
		Status: api.Market_STATUS_OPEN,
	}
	market3 := &api.Market{
		Id:     "c",
		Status: api.Market_STATUS_BETS_LOCKED,
	}
	testCases := []struct {
		desc          string
		req           *api.ListMarketsRequest
		expected      []*api.Market
		expectedCalls int
		expectErr     bool
	}{
		{
			desc:          "basic case",
			req:           &api.ListMarketsRequest{},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListMarketsRequest{PageSize: 1},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListMarketsRequest{PageSize: 2},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListMarketsRequest{PageSize: 3},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListMarketsRequest{PageSize: 4},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "list by status",
			req:           &api.ListMarketsRequest{Status: api.Market_STATUS_OPEN},
			expected:      []*api.Market{market1, market2},
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Markets: []*api.Market{market1, market2, market3}}, log.NewNopLogger())
			var all []*api.Market
			var calls int
			var pageToken string
			for {
				req := proto.Clone(tC.req).(*api.ListMarketsRequest)
				req.PageToken = pageToken
				out, err := s.ListMarkets(context.Background(), connect.NewRequest(req))
				if tC.expectErr {
					require.NotNil(t, err)
					return
				}
				calls++
				require.Nil(t, err)
				require.NotNil(t, out)
				all = append(all, out.Msg.GetMarkets()...)
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

func TestLockMarket(t *testing.T) {
	user := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 100,
	}
	market := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user.Id,
		Status:  api.Market_STATUS_OPEN,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Title: "Yes"},
					{Title: "No"},
				},
			},
		},
	}
	lockedMarket := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user.Id,
		Status:  api.Market_STATUS_BETS_LOCKED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Title: "Yes"},
					{Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc      string
		marketID  string
		expectErr bool
	}{
		{
			desc:     "basic case",
			marketID: market.Id,
		},
		{
			desc:      "fails if market does not exist",
			marketID:  "other",
			expectErr: true,
		},
		{
			desc:      "fails if market is not open",
			marketID:  lockedMarket.Id,
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{market, lockedMarket}}, log.NewNopLogger())
			out, err := s.LockMarket(context.Background(), connect.NewRequest(&api.LockMarketRequest{MarketId: tC.marketID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_BETS_LOCKED, out.Msg.GetMarket().GetStatus())

			got, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: tC.marketID}))
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_BETS_LOCKED, got.Msg.GetMarket().GetStatus())
		})
	}
}

func TestSettleMarket(t *testing.T) {
	user1 := &api.User{
		Id:          uuid.NewString(),
		Username:    "rusty",
		Centipoints: 1000,
	}
	user2 := &api.User{
		Id:          uuid.NewString(),
		Username:    "danny",
		Centipoints: 1000,
	}
	user3 := &api.User{
		Id:          uuid.NewString(),
		Username:    "linus",
		Centipoints: 1000,
	}
	settledMarket := &api.Market{
		Id:      uuid.NewString(),
		Title:   "Will I PB?",
		Creator: user1.Id,
		Status:  api.Market_STATUS_SETTLED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Id: "outcome-1", Title: "Yes"},
					{Id: "outcome-2", Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc                          string
		marketID                      string
		winnerID                      string
		markets                       []*api.Market
		bets                          []*api.Bet
		expectedBetSettledCentipoints map[string]uint64
		expectedUserCentipoints       map[string]uint64
		expectErr                     bool
	}{
		{
			desc:      "fails if market does not exist",
			marketID:  "other",
			winnerID:  "outcome-1",
			expectErr: true,
		},
		{
			desc:      "fails if market is not locked",
			markets:   []*api.Market{settledMarket},
			marketID:  settledMarket.Id,
			winnerID:  "outcome-1",
			expectErr: true,
		},
		{
			desc: "fails if winner does not exist",
			markets: []*api.Market{
				{
					Id:      "z",
					Title:   "Will I PB?",
					Creator: user1.Id,
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Id: "outcome-1", Title: "Yes", Centipoints: 100},
								{Id: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			marketID: "z",
			winnerID: "other",
			bets: []*api.Bet{
				{Id: "a", UserId: user1.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-1"}},
				{Id: "b", UserId: user2.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-2"}},
			},
			expectErr: true,
		},
		{
			markets: []*api.Market{
				{
					Id:      "z",
					Title:   "Will I PB?",
					Creator: user1.Id,
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Id: "outcome-1", Title: "Yes", Centipoints: 100},
								{Id: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			marketID: "z",
			winnerID: "outcome-1",
			bets: []*api.Bet{
				{Id: "a", UserId: user1.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-1"}},
				{Id: "b", UserId: user2.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				"a": 200,
				"b": 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.Id: 1200,
				user2.Id: 1000,
			},
		},
		{
			markets: []*api.Market{
				{
					Id:      "z",
					Title:   "Will I PB?",
					Creator: user1.Id,
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Id: "outcome-1", Title: "Yes", Centipoints: 100},
								{Id: "outcome-2", Title: "No", Centipoints: 150},
							},
						},
					},
				},
			},
			marketID: "z",
			winnerID: "outcome-1",
			bets: []*api.Bet{
				{Id: "a", UserId: user1.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-1"}},
				{Id: "b", UserId: user2.Id, MarketId: "z", Centipoints: 100, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-2"}},
				{Id: "c", UserId: user3.Id, MarketId: "z", Centipoints: 50, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				"a": 250,
				"b": 0,
				"c": 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.Id: 1250,
				user2.Id: 1000,
				user3.Id: 1000,
			},
		},
		{
			markets: []*api.Market{
				{
					Id:      "z",
					Title:   "Will I PB?",
					Creator: user1.Id,
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Id: "outcome-1", Title: "Yes", Centipoints: 100},
								{Id: "outcome-2", Title: "No", Centipoints: 200},
							},
						},
					},
				},
			},
			marketID: "z",
			winnerID: "outcome-1",
			bets: []*api.Bet{
				{Id: "a", UserId: user1.Id, MarketId: "z", Centipoints: 25, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-1"}},
				{Id: "b", UserId: user2.Id, MarketId: "z", Centipoints: 75, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-1"}},
				{Id: "c", UserId: user3.Id, MarketId: "z", Centipoints: 200, Type: &api.Bet_OutcomeId{OutcomeId: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				"a": 75,
				"b": 225,
				"c": 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.Id: 1075,
				user2.Id: 1225,
				user3.Id: 1000,
			},
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Users: []*api.User{proto.Clone(user1).(*api.User), proto.Clone(user2).(*api.User), proto.Clone(user3).(*api.User)}, Markets: tC.markets, Bets: tC.bets}, log.NewNopLogger())
			out, err := s.SettleMarket(context.Background(), connect.NewRequest(&api.SettleMarketRequest{MarketId: tC.marketID, Type: &api.SettleMarketRequest_WinnerId{WinnerId: tC.winnerID}}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, out.Msg.GetMarket().GetStatus(), out.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, out.Msg.GetMarket().GetSettledAt())
			assert.Equal(t, tC.winnerID, out.Msg.GetMarket().GetPool().GetWinnerId())

			got, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: tC.marketID}))
			require.Nil(t, err)
			assert.Equal(t, out.Msg.GetMarket().GetStatus(), got.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, got.Msg.GetMarket().GetSettledAt())
			assert.Equal(t, tC.winnerID, got.Msg.GetMarket().GetPool().GetWinnerId())

			for betID, cp := range tC.expectedBetSettledCentipoints {
				gotBet, err := s.GetBet(context.Background(), connect.NewRequest(&api.GetBetRequest{BetId: betID}))
				require.Nil(t, err)
				assert.NotEmpty(t, gotBet.Msg.GetBet().GetSettledAt())
				assert.Equal(t, cp, gotBet.Msg.GetBet().GetSettledCentipoints(), betID)
			}

			for userID, cp := range tC.expectedUserCentipoints {
				gotUser, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: userID}))
				require.Nil(t, err)
				assert.Equal(t, cp, gotUser.Msg.GetUser().GetCentipoints())
			}
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
		Status:  api.Market_STATUS_OPEN,
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
			s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket, lockedPoolMarket, settledPoolMarket}}, log.NewNopLogger())
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

func TestGetBet(t *testing.T) {
	bet := &api.Bet{
		Id: uuid.NewString(),
	}
	testCases := []struct {
		desc      string
		betID     string
		expected  *api.Bet
		expectErr bool
	}{
		{
			desc:     "basic case",
			betID:    bet.Id,
			expected: bet,
		},
		{
			desc:      "fails if bet does not exist",
			betID:     "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			betID:     "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Bets: []*api.Bet{bet}}, log.NewNopLogger())
			out, err := s.GetBet(context.Background(), connect.NewRequest(&api.GetBetRequest{BetId: tC.betID}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, tC.expected, out.Msg.GetBet())
		})
	}
}

func TestListBets(t *testing.T) {
	// tests pagination until all bets are returned
	// alphabetically ordered ids
	bet1 := &api.Bet{
		Id:       "a",
		UserId:   "rusty",
		MarketId: "one",
	}
	bet2 := &api.Bet{
		Id:       "b",
		UserId:   "danny",
		MarketId: "two",
	}
	bet3 := &api.Bet{
		Id:       "c",
		UserId:   "linus",
		MarketId: "three",
	}
	testCases := []struct {
		desc          string
		req           *api.ListBetsRequest
		expected      []*api.Bet
		expectedCalls int
		expectErr     bool
	}{
		{
			desc:          "basic case",
			req:           &api.ListBetsRequest{},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListBetsRequest{PageSize: 1},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListBetsRequest{PageSize: 2},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListBetsRequest{PageSize: 3},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListBetsRequest{PageSize: 4},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "list by user",
			req:           &api.ListBetsRequest{UserId: "rusty"},
			expected:      []*api.Bet{bet1},
			expectedCalls: 1,
		},
		{
			desc:          "list by market",
			req:           &api.ListBetsRequest{MarketId: "two"},
			expected:      []*api.Bet{bet2},
			expectedCalls: 1,
		},
		{
			desc:          "list by user and market - match",
			req:           &api.ListBetsRequest{UserId: "linus", MarketId: "three"},
			expected:      []*api.Bet{bet3},
			expectedCalls: 1,
		},
		{
			desc:          "list by user and market - no match",
			req:           &api.ListBetsRequest{UserId: "linus", MarketId: "two"},
			expected:      nil,
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s := server.New(&mem.Repo{Bets: []*api.Bet{bet1, bet2, bet3}}, log.NewNopLogger())
			var all []*api.Bet
			var calls int
			var pageToken string
			for {
				req := proto.Clone(tC.req).(*api.ListBetsRequest)
				req.PageToken = pageToken
				out, err := s.ListBets(context.Background(), connect.NewRequest(req))
				if tC.expectErr {
					require.NotNil(t, err)
					return
				}
				calls++
				require.Nil(t, err)
				require.NotNil(t, out)
				all = append(all, out.Msg.GetBets()...)
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

func TestCreateBetConcurrency(t *testing.T) {
	user := &api.User{
		Id:          uuid.NewString(),
		Centipoints: 1000,
		Username:    "rusty",
	}
	poolMarket := &api.Market{
		Id:     uuid.NewString(),
		Status: api.Market_STATUS_OPEN,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Id: uuid.NewString()},
					{Id: uuid.NewString()},
				},
			},
		},
	}
	s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket}}, log.NewNopLogger())
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			_, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{Bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 10,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			}}))
			require.Nil(t, err)
		}()
	}
	wg.Wait()
	m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: poolMarket.Id}))
	require.Nil(t, err)
	assert.Equal(t, uint64(1000), m.Msg.Market.GetPool().GetOutcomes()[0].GetCentipoints())
	assert.Equal(t, uint64(0), m.Msg.Market.GetPool().GetOutcomes()[1].GetCentipoints())
	u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: user.Id}))
	require.Nil(t, err)
	assert.Equal(t, uint64(0), u.Msg.GetUser().GetCentipoints())
}

func TestCreateBetLockMarketConcurrency(t *testing.T) {
	for i := 0; i < 50; i++ {
		user := &api.User{
			Id:          uuid.NewString(),
			Centipoints: 1000,
			Username:    "rusty",
		}
		poolMarket := &api.Market{
			Id:     uuid.NewString(),
			Status: api.Market_STATUS_OPEN,
			Type: &api.Market_Pool{
				Pool: &api.Pool{
					Outcomes: []*api.Outcome{
						{Id: uuid.NewString()},
						{Id: uuid.NewString()},
					},
				},
			},
		}

		s := server.New(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket}}, log.NewNopLogger())
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{Bet: &api.Bet{
				UserId:      user.Id,
				MarketId:    poolMarket.Id,
				Centipoints: 10,
				Type:        &api.Bet_OutcomeId{OutcomeId: poolMarket.GetPool().Outcomes[0].Id},
			}}))
			if err != nil {
				var connectErr *connect.Error
				if errors.As(err, &connectErr) {
					require.NotEqual(t, connect.CodeInternal, connectErr.Code())
					return
				}
			}
		}()
		go func() {
			defer wg.Done()
			_, err := s.LockMarket(context.Background(), connect.NewRequest(&api.LockMarketRequest{MarketId: poolMarket.Id}))
			require.Nil(t, err)
		}()

		wg.Wait()
		m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{MarketId: poolMarket.Id}))
		require.Nil(t, err)
		u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{UserId: user.Id}))
		require.Nil(t, err)

		assert.Equal(t, uint64(1000), u.Msg.GetUser().GetCentipoints()+m.Msg.GetMarket().GetPool().GetOutcomes()[0].GetCentipoints())
	}
}
