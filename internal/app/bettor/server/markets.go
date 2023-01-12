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
