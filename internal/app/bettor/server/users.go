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

// CreateUser creates a new user.
func (s *Server) CreateUser(ctx context.Context, in *connect.Request[api.CreateUserRequest]) (*connect.Response[api.CreateUserResponse], error) {
	if in.Msg == nil || in.Msg.GetUser() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user is required"))
	}
	user := proto.Clone(in.Msg.GetUser()).(*api.User)

	user.Id = uuid.NewString()
	user.CreatedAt = timestamppb.Now()
	user.UpdatedAt = timestamppb.Now()

	if err := user.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if err := s.Repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.CreateUserResponse{
		User: user,
	}), nil
}

// GetUser returns a user by ID.
func (s *Server) GetUser(ctx context.Context, in *connect.Request[api.GetUserRequest]) (*connect.Response[api.GetUserResponse], error) {
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	user, err := s.Repo.GetUser(ctx, in.Msg.GetUserId())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetUserResponse{
		User: user,
	}), nil
}

// GetUserByUsername returns a user by username.
func (s *Server) GetUserByUsername(ctx context.Context, in *connect.Request[api.GetUserByUsernameRequest]) (*connect.Response[api.GetUserByUsernameResponse], error) {
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	user, err := s.Repo.GetUserByUsername(ctx, in.Msg.GetUsername())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetUserByUsernameResponse{
		User: user,
	}), nil
}
