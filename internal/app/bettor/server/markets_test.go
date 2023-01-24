package server_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreateMarket(t *testing.T) {
	var maxOpenMarkets []*api.Market
	for i := 0; i < server.MaxNumberOfOpenMarkets; i++ {
		maxOpenMarkets = append(maxOpenMarkets, &api.Market{
			Name:   entity.MarketN("guild:A", uuid.NewString()),
			Status: api.Market_STATUS_OPEN,
		})
	}
	user := &api.User{
		Name:        entity.UserN("guild:A", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	userB := &api.User{
		Name:        entity.UserN("guild:B", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	testCases := []struct {
		desc            string
		existingMarkets []*api.Market
		book            string
		market          *api.Market
		expectErr       bool
	}{
		{
			desc: "basic case",
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
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
			desc: "fails if book not set",
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
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
			desc: "fails if creator not in the same book as market",
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: userB.GetName(),
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
			desc: "fails if title not set",
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Creator: user.GetName(),
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
			book: entity.BookN("guild:A"),
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
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
			},
			expectErr: true,
		},
		{
			desc: "fails if pool has less than 2 outcomes",
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
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
			desc: "fails if duplicate outcome titles",
			book: entity.BookN("guild:A"),
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: []*api.Outcome{
							{Title: "Yes"},
							{Title: "Yes"},
						},
					},
				},
			},
			expectErr: true,
		},
		{
			desc:            "fails if max number of open markets reached",
			book:            entity.BookN("guild:A"),
			existingMarkets: maxOpenMarkets,
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: user.GetName(),
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
			desc:            "only enforces max number of open markets in the same book",
			book:            entity.BookN("guild:B"),
			existingMarkets: maxOpenMarkets,
			market: &api.Market{
				Title:   "Will I PB?",
				Creator: userB.GetName(),
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
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user, userB}, Markets: tC.existingMarkets}))
			require.Nil(t, err)
			out, err := s.CreateMarket(context.Background(), connect.NewRequest(&api.CreateMarketRequest{Book: tC.book, Market: tC.market}))
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
		Name: entity.MarketN("guild:1", uuid.NewString()),
	}
	testCases := []struct {
		desc      string
		market    string
		expected  *api.Market
		expectErr bool
	}{
		{
			desc:     "basic case",
			market:   market.GetName(),
			expected: market,
		},
		{
			desc:      "fails if market does not exist",
			market:    "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			market:    "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Markets: []*api.Market{market}}))
			require.Nil(t, err)
			out, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: tC.market}))
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
		Name:   entity.MarketN("guild:1", "a"),
		Status: api.Market_STATUS_OPEN,
	}
	market2 := &api.Market{
		Name:   entity.MarketN("guild:1", "b"),
		Status: api.Market_STATUS_OPEN,
	}
	market3 := &api.Market{
		Name:   entity.MarketN("guild:1", "c"),
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
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1")},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "searches within book",
			req:           &api.ListMarketsRequest{Book: entity.BookN("other")},
			expected:      nil,
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1"), PageSize: 1},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1"), PageSize: 2},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1"), PageSize: 3},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1"), PageSize: 4},
			expected:      []*api.Market{market1, market2, market3},
			expectedCalls: 1,
		},
		{
			desc:          "list by status",
			req:           &api.ListMarketsRequest{Book: entity.BookN("guild:1"), Status: api.Market_STATUS_OPEN},
			expected:      []*api.Market{market1, market2},
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Markets: []*api.Market{market1, market2, market3}}))
			require.Nil(t, err)
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
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 100,
	}
	market := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user.GetName(),
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
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user.GetName(),
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
		market    string
		expectErr bool
	}{
		{
			desc:   "basic case",
			market: market.GetName(),
		},
		{
			desc:      "fails if market does not exist",
			market:    "other",
			expectErr: true,
		},
		{
			desc:      "fails if market is not open",
			market:    lockedMarket.GetName(),
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{market, lockedMarket}}))
			require.Nil(t, err)
			out, err := s.LockMarket(context.Background(), connect.NewRequest(&api.LockMarketRequest{Name: tC.market}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_BETS_LOCKED, out.Msg.GetMarket().GetStatus())

			got, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: tC.market}))
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_BETS_LOCKED, got.Msg.GetMarket().GetStatus())
		})
	}
}

func TestSettleMarket(t *testing.T) {
	marketName := entity.MarketN("guild:1", uuid.NewString())
	user1 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 1000,
	}
	user2 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "danny",
		Centipoints: 1000,
	}
	user3 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "linus",
		Centipoints: 1000,
	}
	settledMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user1.GetName(),
		Status:  api.Market_STATUS_SETTLED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: "outcome-1", Title: "Yes"},
					{Name: "outcome-2", Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc                          string
		market                        string
		winner                        string
		markets                       []*api.Market
		bets                          []*api.Bet
		expectedBetSettledCentipoints map[string]uint64
		expectedUserCentipoints       map[string]uint64
		expectErr                     bool
	}{
		{
			desc:      "fails if market does not exist",
			market:    marketName,
			winner:    "outcome-1",
			expectErr: true,
		},
		{
			desc:      "fails if market is not locked",
			markets:   []*api.Market{settledMarket},
			market:    settledMarket.GetName(),
			winner:    "outcome-1",
			expectErr: true,
		},
		{
			desc: "fails if winner does not exist",
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			market: marketName,
			winner: "other",
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectErr: true,
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			market: marketName,
			winner: "outcome-1",
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 200,
				entity.BetN("guild:1", "b"): 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1200,
				user2.GetName(): 1000,
			},
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 150},
							},
						},
					},
				},
			},
			market: marketName,
			winner: "outcome-1",
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
				{Name: entity.BetN("guild:1", "c"), User: user3.GetName(), Market: marketName, Centipoints: 50, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 250,
				entity.BetN("guild:1", "b"): 0,
				entity.BetN("guild:1", "c"): 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1250,
				user2.GetName(): 1000,
				user3.GetName(): 1000,
			},
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 200},
							},
						},
					},
				},
			},
			market: marketName,
			winner: "outcome-1",
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 25, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 75, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "c"), User: user3.GetName(), Market: marketName, Centipoints: 200, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 75,
				entity.BetN("guild:1", "b"): 225,
				entity.BetN("guild:1", "c"): 0,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1075,
				user2.GetName(): 1225,
				user3.GetName(): 1000,
			},
		},
		{
			desc: "nop if there were no bets",
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			market:                        marketName,
			winner:                        "outcome-1",
			bets:                          []*api.Bet{},
			expectedBetSettledCentipoints: map[string]uint64{},
			expectedUserCentipoints:       map[string]uint64{},
		},
		{
			desc: "refund bets if there are no winning bets",
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 0},
								{Name: "outcome-2", Title: "No", Centipoints: 150},
							},
						},
					},
				},
			},
			market: marketName,
			winner: "outcome-1",
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 50, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 100,
				entity.BetN("guild:1", "b"): 50,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1100,
				user2.GetName(): 1050,
			},
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{proto.Clone(user1).(*api.User), proto.Clone(user2).(*api.User), proto.Clone(user3).(*api.User)}, Markets: tC.markets, Bets: tC.bets}))
			require.Nil(t, err)
			out, err := s.SettleMarket(context.Background(), connect.NewRequest(&api.SettleMarketRequest{Name: tC.market, Type: &api.SettleMarketRequest_Winner{Winner: tC.winner}}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_SETTLED, out.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, out.Msg.GetMarket().GetSettledAt())
			assert.Equal(t, tC.winner, out.Msg.GetMarket().GetPool().GetWinner())

			got, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: tC.market}))
			require.Nil(t, err)
			assert.Equal(t, out.Msg.GetMarket().GetStatus(), got.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, got.Msg.GetMarket().GetSettledAt())
			assert.Equal(t, tC.winner, got.Msg.GetMarket().GetPool().GetWinner())

			for betN, cp := range tC.expectedBetSettledCentipoints {
				gotBet, err := s.GetBet(context.Background(), connect.NewRequest(&api.GetBetRequest{Bet: betN}))
				require.Nil(t, err)
				assert.NotEmpty(t, gotBet.Msg.GetBet().GetSettledAt())
				assert.Equal(t, cp, gotBet.Msg.GetBet().GetSettledCentipoints(), betN)
			}

			for userN, cp := range tC.expectedUserCentipoints {
				gotUser, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: userN}))
				require.Nil(t, err)
				assert.Equal(t, cp, gotUser.Msg.GetUser().GetCentipoints())
			}
		})
	}
}

func TestCancelMarket(t *testing.T) {
	marketName := entity.MarketN("guild:1", uuid.NewString())
	user1 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 1000,
	}
	user2 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "danny",
		Centipoints: 1000,
	}
	user3 := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "linus",
		Centipoints: 1000,
	}
	settledMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user1.GetName(),
		Status:  api.Market_STATUS_SETTLED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: "outcome-1", Title: "Yes"},
					{Name: "outcome-2", Title: "No"},
				},
			},
		},
	}
	canceledMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user1.GetName(),
		Status:  api.Market_STATUS_CANCELED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: "outcome-1", Title: "Yes"},
					{Name: "outcome-2", Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc                          string
		market                        string
		markets                       []*api.Market
		bets                          []*api.Bet
		expectedBetSettledCentipoints map[string]uint64
		expectedUserCentipoints       map[string]uint64
		expectErr                     bool
	}{
		{
			desc:      "fails if market does not exist",
			market:    marketName,
			expectErr: true,
		},
		{
			desc:      "fails if market is already settled",
			markets:   []*api.Market{settledMarket},
			market:    settledMarket.GetName(),
			expectErr: true,
		},
		{
			desc:      "fails if market is already canceled",
			markets:   []*api.Market{canceledMarket},
			market:    canceledMarket.GetName(),
			expectErr: true,
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			market: marketName,
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 100,
				entity.BetN("guild:1", "b"): 100,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1100,
				user2.GetName(): 1100,
			},
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 150},
							},
						},
					},
				},
			},
			market: marketName,
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 100, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
				{Name: entity.BetN("guild:1", "c"), User: user3.GetName(), Market: marketName, Centipoints: 50, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 100,
				entity.BetN("guild:1", "b"): 100,
				entity.BetN("guild:1", "c"): 50,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1100,
				user2.GetName(): 1100,
				user3.GetName(): 1050,
			},
		},
		{
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 200},
							},
						},
					},
				},
			},
			market: marketName,
			bets: []*api.Bet{
				{Name: entity.BetN("guild:1", "a"), User: user1.GetName(), Market: marketName, Centipoints: 25, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "b"), User: user2.GetName(), Market: marketName, Centipoints: 75, Type: &api.Bet_Outcome{Outcome: "outcome-1"}},
				{Name: entity.BetN("guild:1", "c"), User: user3.GetName(), Market: marketName, Centipoints: 200, Type: &api.Bet_Outcome{Outcome: "outcome-2"}},
			},
			expectedBetSettledCentipoints: map[string]uint64{
				entity.BetN("guild:1", "a"): 25,
				entity.BetN("guild:1", "b"): 75,
				entity.BetN("guild:1", "c"): 200,
			},
			expectedUserCentipoints: map[string]uint64{
				user1.GetName(): 1025,
				user2.GetName(): 1075,
				user3.GetName(): 1200,
			},
		},
		{
			desc: "nop if there were no bets",
			markets: []*api.Market{
				{
					Name:    marketName,
					Title:   "Will I PB?",
					Creator: user1.GetName(),
					Status:  api.Market_STATUS_BETS_LOCKED,
					Type: &api.Market_Pool{
						Pool: &api.Pool{
							Outcomes: []*api.Outcome{
								{Name: "outcome-1", Title: "Yes", Centipoints: 100},
								{Name: "outcome-2", Title: "No", Centipoints: 100},
							},
						},
					},
				},
			},
			market:                        marketName,
			bets:                          []*api.Bet{},
			expectedBetSettledCentipoints: map[string]uint64{},
			expectedUserCentipoints:       map[string]uint64{},
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{proto.Clone(user1).(*api.User), proto.Clone(user2).(*api.User), proto.Clone(user3).(*api.User)}, Markets: tC.markets, Bets: tC.bets}))
			require.Nil(t, err)
			out, err := s.CancelMarket(context.Background(), connect.NewRequest(&api.CancelMarketRequest{Name: tC.market}))
			if tC.expectErr {
				fmt.Println(err)
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, api.Market_STATUS_CANCELED, out.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, out.Msg.GetMarket().GetSettledAt())

			got, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: tC.market}))
			require.Nil(t, err)
			assert.Equal(t, out.Msg.GetMarket().GetStatus(), got.Msg.GetMarket().GetStatus())
			assert.NotEmpty(t, got.Msg.GetMarket().GetSettledAt())

			for betN, cp := range tC.expectedBetSettledCentipoints {
				gotBet, err := s.GetBet(context.Background(), connect.NewRequest(&api.GetBetRequest{Bet: betN}))
				require.Nil(t, err)
				assert.NotEmpty(t, gotBet.Msg.GetBet().GetSettledAt())
				assert.Equal(t, cp, gotBet.Msg.GetBet().GetSettledCentipoints(), betN)
			}

			for userN, cp := range tC.expectedUserCentipoints {
				gotUser, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: userN}))
				require.Nil(t, err)
				assert.Equal(t, cp, gotUser.Msg.GetUser().GetCentipoints())
			}
		})
	}
}

func TestCreateBet(t *testing.T) {
	user := &api.User{
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 1000,
	}
	user2 := &api.User{
		Name:        entity.UserN("guild:2", uuid.NewString()),
		Username:    "rusty",
		Centipoints: 1000,
	}
	poolMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user.GetName(),
		Status:  api.Market_STATUS_OPEN,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: uuid.NewString(), Title: "Yes"},
					{Name: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	lockedPoolMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user.GetName(),
		Status:  api.Market_STATUS_BETS_LOCKED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: uuid.NewString(), Title: "Yes"},
					{Name: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	settledPoolMarket := &api.Market{
		Name:    entity.MarketN("guild:1", uuid.NewString()),
		Title:   "Will I PB?",
		Creator: user.GetName(),
		Status:  api.Market_STATUS_SETTLED,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: uuid.NewString(), Title: "Yes"},
					{Name: uuid.NewString(), Title: "No"},
				},
			},
		},
	}
	testCases := []struct {
		desc                  string
		book                  string
		bet                   *api.Bet
		expectUserCentipoints uint64
		expectErr             bool
	}{
		// pool bets
		{
			desc: "basic case - pool bet",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectUserCentipoints: 900,
		},
		{
			desc: "fails if user is not in the same book as bet",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user2.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if market is not in the same book as bet",
			book: entity.BookN("guild:other"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectUserCentipoints: 900,
			expectErr:             true,
		},
		{
			desc: "fails if book not set",
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if user does not exist",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        "other",
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if market does not exist",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      "other",
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if type not provided",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
			},
			expectErr: true,
		},
		{
			desc: "fails if outcome does not exist",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: "other"},
			},
			expectErr: true,
		},
		{
			desc: "fails if creating a bet on a locked market",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      lockedPoolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: lockedPoolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if creating a bet on a settled market",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      settledPoolMarket.GetName(),
				Centipoints: 100,
				Type:        &api.Bet_Outcome{Outcome: settledPoolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
		{
			desc: "fails if betting more points than user has",
			book: entity.BookN("guild:1"),
			bet: &api.Bet{
				User:        user.GetName(),
				Market:      poolMarket.GetName(),
				Centipoints: 2000,
				Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
			},
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user, user2}, Markets: []*api.Market{poolMarket, lockedPoolMarket, settledPoolMarket}}))
			require.Nil(t, err)
			out, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{Book: tC.book, Bet: tC.bet}))
			if tC.expectErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)

			assert.NotEmpty(t, out)

			u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: tC.bet.GetUser()}))
			require.Nil(t, err)
			assert.Equal(t, tC.expectUserCentipoints, u.Msg.User.GetCentipoints())

			m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: tC.bet.GetMarket()}))
			require.Nil(t, err)
			assert.Equal(t, tC.bet.GetCentipoints(), m.Msg.Market.GetPool().GetOutcomes()[0].GetCentipoints())
		})
	}
}

func TestGetBet(t *testing.T) {
	bet := &api.Bet{
		Name: entity.BetN("guild:1", uuid.NewString()),
	}
	testCases := []struct {
		desc      string
		bet       string
		expected  *api.Bet
		expectErr bool
	}{
		{
			desc:     "basic case",
			bet:      bet.GetName(),
			expected: bet,
		},
		{
			desc:      "fails if bet does not exist",
			bet:       "does-not-exist",
			expectErr: true,
		},
		{
			desc:      "fails if id is empty",
			bet:       "",
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Bets: []*api.Bet{bet}}))
			require.Nil(t, err)
			out, err := s.GetBet(context.Background(), connect.NewRequest(&api.GetBetRequest{Bet: tC.bet}))
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
		Name:   entity.BetN("guild:1", "a"),
		User:   "rusty",
		Market: "one",
	}
	bet2 := &api.Bet{
		Name:   entity.BetN("guild:1", "b"),
		User:   "danny",
		Market: "two",
	}
	bet3 := &api.Bet{
		Name:      entity.BetN("guild:1", "c"),
		User:      "linus",
		Market:    "three",
		SettledAt: timestamppb.Now(),
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
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1")},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "list within book",
			req:           &api.ListBetsRequest{Book: "other"},
			expected:      nil,
			expectedCalls: 1,
		},
		{
			desc:          "page size 1",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), PageSize: 1},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 3,
		},
		{
			desc:          "page size 2",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), PageSize: 2},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 2,
		},
		{
			desc:          "page size 3",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), PageSize: 3},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "page size 4",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), PageSize: 4},
			expected:      []*api.Bet{bet1, bet2, bet3},
			expectedCalls: 1,
		},
		{
			desc:          "list by user",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), User: "rusty"},
			expected:      []*api.Bet{bet1},
			expectedCalls: 1,
		},
		{
			desc:          "list by market",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), Market: "two"},
			expected:      []*api.Bet{bet2},
			expectedCalls: 1,
		},
		{
			desc:          "list excluding settled",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), ExcludeSettled: true},
			expected:      []*api.Bet{bet1, bet2},
			expectedCalls: 1,
		},
		{
			desc:          "list by user and market - match",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), User: "linus", Market: "three"},
			expected:      []*api.Bet{bet3},
			expectedCalls: 1,
		},
		{
			desc:          "list by user and market - no match",
			req:           &api.ListBetsRequest{Book: entity.BookN("guild:1"), User: "linus", Market: "two"},
			expected:      nil,
			expectedCalls: 1,
		},
	}
	for _, tC := range testCases {
		tC := tC
		t.Run(tC.desc, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{Bets: []*api.Bet{bet1, bet2, bet3}}))
			require.Nil(t, err)
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
		Name:        entity.UserN("guild:1", uuid.NewString()),
		Centipoints: 1000,
		Username:    "rusty",
	}
	poolMarket := &api.Market{
		Name:   entity.MarketN("guild:1", uuid.NewString()),
		Status: api.Market_STATUS_OPEN,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: []*api.Outcome{
					{Name: uuid.NewString()},
					{Name: uuid.NewString()},
				},
			},
		},
	}
	s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket}}))
	require.Nil(t, err)
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			_, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{
				Book: entity.BookN("guild:1"),
				Bet: &api.Bet{
					User:        user.GetName(),
					Market:      poolMarket.GetName(),
					Centipoints: 10,
					Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
				},
			}))
			require.Nil(t, err)
		}()
	}
	wg.Wait()
	m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: poolMarket.GetName()}))
	require.Nil(t, err)
	assert.Equal(t, uint64(1000), m.Msg.Market.GetPool().GetOutcomes()[0].GetCentipoints())
	assert.Equal(t, uint64(0), m.Msg.Market.GetPool().GetOutcomes()[1].GetCentipoints())
	u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: user.GetName()}))
	require.Nil(t, err)
	assert.Equal(t, uint64(0), u.Msg.GetUser().GetCentipoints())
}

func TestCreateBetLockMarketConcurrency(t *testing.T) {
	for i := 0; i < 50; i++ {
		user := &api.User{
			Name:        entity.UserN("guild:1", uuid.NewString()),
			Centipoints: 1000,
			Username:    "rusty",
		}
		poolMarket := &api.Market{
			Name:   entity.MarketN("guild:1", uuid.NewString()),
			Status: api.Market_STATUS_OPEN,
			Type: &api.Market_Pool{
				Pool: &api.Pool{
					Outcomes: []*api.Outcome{
						{Name: uuid.NewString()},
						{Name: uuid.NewString()},
					},
				},
			},
		}

		s, err := server.New(server.WithRepo(&mem.Repo{Users: []*api.User{user}, Markets: []*api.Market{poolMarket}}))
		require.Nil(t, err)
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := s.CreateBet(context.Background(), connect.NewRequest(&api.CreateBetRequest{
				Book: entity.BookN("guild:1"),
				Bet: &api.Bet{
					User:        user.GetName(),
					Market:      poolMarket.GetName(),
					Centipoints: 10,
					Type:        &api.Bet_Outcome{Outcome: poolMarket.GetPool().Outcomes[0].GetName()},
				},
			}))
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
			_, err := s.LockMarket(context.Background(), connect.NewRequest(&api.LockMarketRequest{Name: poolMarket.GetName()}))
			require.Nil(t, err)
		}()

		wg.Wait()
		m, err := s.GetMarket(context.Background(), connect.NewRequest(&api.GetMarketRequest{Name: poolMarket.GetName()}))
		require.Nil(t, err)
		u, err := s.GetUser(context.Background(), connect.NewRequest(&api.GetUserRequest{Name: user.GetName()}))
		require.Nil(t, err)

		assert.Equal(t, uint64(1000), u.Msg.GetUser().GetCentipoints()+m.Msg.GetMarket().GetPool().GetOutcomes()[0].GetCentipoints())
	}
}
