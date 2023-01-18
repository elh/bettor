package server

import (
	"context"
	"encoding/gob"
	"errors"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/entity"
	"github.com/elh/bettor/internal/app/bettor/repo"
	"github.com/elh/bettor/internal/pkg/pagination"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MaxNumberOfOpenMarkets is the maximum number of open markets allowed.
const MaxNumberOfOpenMarkets = 25

func init() {
	gob.Register(&api.ListMarketsRequest{})
	gob.Register(&api.ListBetsRequest{})
}

// CreateMarket creates a new betting market.
func (s *Server) CreateMarket(ctx context.Context, in *connect.Request[api.CreateMarketRequest]) (*connect.Response[api.CreateMarketResponse], error) {
	s.marketMtx.RLock()
	defer s.marketMtx.RUnlock()
	if in.Msg == nil || in.Msg.GetMarket() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is required"))
	}
	market := proto.Clone(in.Msg.GetMarket()).(*api.Market)

	marketID := uuid.NewString()
	market.Name = entity.MarketN(marketID)
	market.CreatedAt = timestamppb.Now()
	market.UpdatedAt = timestamppb.Now()
	market.SettledAt = nil
	market.Status = api.Market_STATUS_OPEN

	if market.GetPool() != nil {
		market.GetPool().Winner = ""
		outcomeTitles := map[string]bool{}
		for _, outcome := range market.GetPool().GetOutcomes() {
			outcome.Name = entity.OutcomeN(marketID, uuid.NewString())
			outcome.Centipoints = 0
			if outcomeTitles[outcome.GetTitle()] {
				return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("duplicate outcome title"))
			}
			outcomeTitles[outcome.GetTitle()] = true
		}
	}

	openMarkets, _, err := s.Repo.ListMarkets(ctx, &repo.ListMarketsArgs{
		Status: api.Market_STATUS_OPEN,
		Limit:  MaxNumberOfOpenMarkets,
	})
	if err != nil {
		return nil, err
	}
	if len(openMarkets) >= MaxNumberOfOpenMarkets {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("too many open markets"))
	}

	if _, err := s.Repo.GetUser(ctx, market.GetCreator()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := market.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.Repo.CreateMarket(ctx, market); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.CreateMarketResponse{
		Market: market,
	}), nil
}

// GetMarket returns a market by ID.
func (s *Server) GetMarket(ctx context.Context, in *connect.Request[api.GetMarketRequest]) (*connect.Response[api.GetMarketResponse], error) {
	s.marketMtx.RLock()
	defer s.marketMtx.RUnlock()
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	market, err := s.Repo.GetMarket(ctx, in.Msg.GetName())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetMarketResponse{
		Market: market,
	}), nil
}

// ListMarkets lists markets by filters.
func (s *Server) ListMarkets(ctx context.Context, in *connect.Request[api.ListMarketsRequest]) (*connect.Response[api.ListMarketsResponse], error) {
	pageSize := defaultPageSize
	if in.Msg.GetPageSize() > 0 && in.Msg.GetPageSize() <= maxPageSize {
		pageSize = int(in.Msg.GetPageSize())
	}

	var cursor string
	if in.Msg.GetPageToken() != "" {
		p, err := pagination.FromToken(in.Msg.GetPageToken())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		cursor = p.Cursor
		fromToken, ok := proto.Clone(p.ListRequest).(*api.ListMarketsRequest)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
		if !proto.Equal(api.StripListMarketsPagination(in.Msg), api.StripListMarketsPagination(fromToken)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
	}

	markets, hasMore, err := s.Repo.ListMarkets(ctx, &repo.ListMarketsArgs{
		GreaterThanID: cursor,
		Status:        in.Msg.GetStatus(),
		Limit:         pageSize,
	})
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if hasMore {
		nextPageToken, err = pagination.ToToken(pagination.Pagination{
			Cursor:      markets[len(markets)-1].GetName(),
			ListRequest: in.Msg,
		})
		if err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&api.ListMarketsResponse{
		Markets:       markets,
		NextPageToken: nextPageToken,
	}), nil
}

// SettleMarket settles a betting market and pays out bets.
func (s *Server) SettleMarket(ctx context.Context, in *connect.Request[api.SettleMarketRequest]) (*connect.Response[api.SettleMarketResponse], error) {
	s.marketMtx.Lock()
	defer s.marketMtx.Unlock()
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	market, err := s.Repo.GetMarket(ctx, in.Msg.GetName())
	if err != nil {
		return nil, err
	}
	if market.GetStatus() != api.Market_STATUS_BETS_LOCKED {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is not locked"))
	}
	market.Status = api.Market_STATUS_SETTLED
	market.UpdatedAt = timestamppb.Now()
	market.SettledAt = timestamppb.Now()

	// NOTE: only Pool is supported right now
	if in.Msg.GetWinner() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("winner is required"))
	}

	if market.GetPool() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market has no pool"))
	}
	var found bool
	for _, outcome := range market.GetPool().GetOutcomes() {
		if outcome.GetName() == in.Msg.GetWinner() {
			found = true
			break
		}
	}
	if !found {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("winner is not in pool"))
	}
	market.GetPool().Winner = in.Msg.GetWinner()

	// compute return ratio
	var totalCentipointsBet, winnerCentipointsBet uint64
	for _, outcome := range market.GetPool().GetOutcomes() {
		totalCentipointsBet += outcome.GetCentipoints()
		if outcome.GetName() == in.Msg.GetWinner() {
			winnerCentipointsBet = outcome.GetCentipoints()
		}
	}
	if totalCentipointsBet > 0 {
		winnerRatio := float64(totalCentipointsBet) / float64(winnerCentipointsBet)

		var bets []*api.Bet
		var greaterThanID string
		for {
			bs, hasMore, err := s.Repo.ListBets(ctx, &repo.ListBetsArgs{GreaterThanID: greaterThanID, Market: market.GetName(), Limit: 100})
			if err != nil {
				return nil, err
			}
			bets = append(bets, bs...)
			if !hasMore {
				break
			}
			greaterThanID = bs[len(bs)-1].GetName()
		}
		var hasWinner bool
		for _, bet := range bets {
			if bet.GetOutcome() == market.GetPool().GetWinner() {
				hasWinner = true
				break
			}
		}

		for _, bet := range bets {
			bet.UpdatedAt = timestamppb.Now()
			bet.SettledAt = timestamppb.Now()
			if hasWinner {
				if bet.GetOutcome() == market.GetPool().GetWinner() {
					bet.SettledCentipoints = uint64(float64(bet.GetCentipoints()) * winnerRatio)
				}
			} else {
				bet.SettledCentipoints = bet.GetCentipoints()
			}
		}

		for _, bet := range bets {
			user, err := s.Repo.GetUser(ctx, bet.GetUser())
			if err != nil {
				return nil, err
			}
			user.UpdatedAt = timestamppb.Now()
			user.Centipoints += bet.GetSettledCentipoints()

			if err := s.Repo.UpdateUser(ctx, user); err != nil {
				return nil, err
			}

			if err := s.Repo.UpdateBet(ctx, bet); err != nil {
				return nil, err
			}
		}
	}

	if err := s.Repo.UpdateMarket(ctx, market); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.SettleMarketResponse{Market: market}), nil
}

// LockMarket locks a betting market preventing further bets.
func (s *Server) LockMarket(ctx context.Context, in *connect.Request[api.LockMarketRequest]) (*connect.Response[api.LockMarketResponse], error) {
	s.marketMtx.Lock()
	defer s.marketMtx.Unlock()
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	market, err := s.Repo.GetMarket(ctx, in.Msg.GetName())
	if err != nil {
		return nil, err
	}
	if market.GetStatus() != api.Market_STATUS_OPEN {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is not open"))
	}
	market.Status = api.Market_STATUS_BETS_LOCKED
	market.UpdatedAt = timestamppb.Now()

	if err := s.Repo.UpdateMarket(ctx, market); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.LockMarketResponse{Market: market}), nil
}

// CreateBet places a bet on an open betting market.
func (s *Server) CreateBet(ctx context.Context, in *connect.Request[api.CreateBetRequest]) (*connect.Response[api.CreateBetResponse], error) {
	s.marketMtx.Lock()
	defer s.marketMtx.Unlock()
	if in.Msg == nil || in.Msg.GetBet() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bet is required"))
	}
	bet := proto.Clone(in.Msg.GetBet()).(*api.Bet)

	bet.Name = entity.BetN(uuid.NewString())
	bet.CreatedAt = timestamppb.Now()
	bet.UpdatedAt = timestamppb.Now()
	bet.SettledAt = nil
	bet.SettledCentipoints = 0

	user, err := s.Repo.GetUser(ctx, bet.GetUser())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if user.GetCentipoints() < bet.GetCentipoints() {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user does not have enough balance"))
	}

	market, err := s.Repo.GetMarket(ctx, bet.GetMarket())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if market.GetStatus() != api.Market_STATUS_OPEN {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is not open"))
	}
	if bet.Type == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bet type is required"))
	}
	if bet.GetOutcome() != "" {
		if market.GetPool() == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market does not have a pool"))
		}
		found := false
		for _, outcome := range market.GetPool().GetOutcomes() {
			if outcome.GetName() == bet.GetOutcome() {
				found = true
				break
			}
		}
		if !found {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("outcome not found in market"))
		}
	}

	if err := bet.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// writes
	if err := s.Repo.CreateBet(ctx, bet); err != nil {
		return nil, err
	}
	user.Centipoints -= bet.GetCentipoints()
	if err := s.Repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	if bet.GetOutcome() != "" && market.GetPool() != nil {
		for _, outcome := range market.GetPool().GetOutcomes() {
			if outcome.GetName() == bet.GetOutcome() {
				outcome.Centipoints += bet.GetCentipoints()
				break
			}
		}
	}
	if err := s.Repo.UpdateMarket(ctx, market); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.CreateBetResponse{
		Bet: bet,
	}), nil
}

// GetBet returns a bet by ID.
func (s *Server) GetBet(ctx context.Context, in *connect.Request[api.GetBetRequest]) (*connect.Response[api.GetBetResponse], error) {
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	bet, err := s.Repo.GetBet(ctx, in.Msg.GetBet())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetBetResponse{
		Bet: bet,
	}), nil
}

// ListBets lists bets by filters.
func (s *Server) ListBets(ctx context.Context, in *connect.Request[api.ListBetsRequest]) (*connect.Response[api.ListBetsResponse], error) {
	pageSize := defaultPageSize
	if in.Msg.GetPageSize() > 0 && in.Msg.GetPageSize() <= maxPageSize {
		pageSize = int(in.Msg.GetPageSize())
	}

	var cursor string
	if in.Msg.GetPageToken() != "" {
		p, err := pagination.FromToken(in.Msg.GetPageToken())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		cursor = p.Cursor
		fromToken, ok := proto.Clone(p.ListRequest).(*api.ListBetsRequest)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
		if !proto.Equal(api.StripListBetsPagination(in.Msg), api.StripListBetsPagination(fromToken)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
	}

	bets, hasMore, err := s.Repo.ListBets(ctx, &repo.ListBetsArgs{
		GreaterThanID:  cursor,
		User:           in.Msg.GetUser(),
		Market:         in.Msg.GetMarket(),
		ExcludeSettled: in.Msg.GetExcludeSettled(),
		Limit:          pageSize,
	})
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if hasMore {
		nextPageToken, err = pagination.ToToken(pagination.Pagination{
			Cursor:      bets[len(bets)-1].GetName(),
			ListRequest: in.Msg,
		})
		if err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&api.ListBetsResponse{
		Bets:          bets,
		NextPageToken: nextPageToken,
	}), nil
}
