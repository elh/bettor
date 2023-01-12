package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateMarket creates a new betting market.
func (s *Server) CreateMarket(ctx context.Context, in *connect.Request[api.CreateMarketRequest]) (*connect.Response[api.CreateMarketResponse], error) {
	if in.Msg == nil || in.Msg.GetMarket() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is required"))
	}
	market := proto.Clone(in.Msg.GetMarket()).(*api.Market)

	market.Id = uuid.NewString()
	market.CreatedAt = timestamppb.Now()
	market.UpdatedAt = timestamppb.Now()
	market.LockAt = nil
	market.SettledAt = nil
	market.Status = api.Market_STATUS_ACTIVE

	if market.GetPool() != nil {
		market.GetPool().WinnerId = ""
		for _, outcome := range market.GetPool().GetOutcomes() {
			outcome.Id = uuid.NewString()
			outcome.Centipoints = 0
			outcome.UserCount = 0
		}
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
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	market, err := s.Repo.GetMarket(ctx, in.Msg.GetMarketId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetMarketResponse{
		Market: market,
	}), nil
}

// CreateBet places a bet on an active betting market.
func (s *Server) CreateBet(ctx context.Context, in *connect.Request[api.CreateBetRequest]) (*connect.Response[api.CreateBetResponse], error) {
	if in.Msg == nil || in.Msg.GetBet() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("bet is required"))
	}
	bet := proto.Clone(in.Msg.GetBet()).(*api.Bet)

	bet.Id = uuid.NewString()
	bet.CreatedAt = timestamppb.Now()
	bet.UpdatedAt = timestamppb.Now()
	bet.SettledAt = nil
	bet.Centipoints = 0
	bet.SettledCentipoints = 0

	if _, err := s.Repo.GetUser(ctx, bet.GetUserId()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	market, err := s.Repo.GetMarket(ctx, bet.GetMarketId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if market.GetStatus() != api.Market_STATUS_ACTIVE {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market is not active"))
	}
	if bet.GetOutcomeId() != "" {
		if market.GetPool() == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("market does not have a pool"))
		}
		found := false
		for _, outcome := range market.GetPool().GetOutcomes() {
			if outcome.GetId() == bet.GetOutcomeId() {
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

	if err := s.Repo.CreateBet(ctx, bet); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.CreateBetResponse{
		Bet: bet,
	}), nil
}
