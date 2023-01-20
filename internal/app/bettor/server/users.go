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

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

func init() {
	gob.Register(&api.ListUsersRequest{})
}

// CreateUser creates a new user.
func (s *Server) CreateUser(ctx context.Context, in *connect.Request[api.CreateUserRequest]) (*connect.Response[api.CreateUserResponse], error) {
	if in.Msg == nil || in.Msg.GetUser() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user is required"))
	}
	if in.Msg.GetBook() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("book is required"))
	}
	user := proto.Clone(in.Msg.GetUser()).(*api.User)

	user.Name = entity.UserN(in.Msg.GetBook(), uuid.NewString())
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

	user, err := s.Repo.GetUser(ctx, in.Msg.GetName())
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

	user, err := s.Repo.GetUserByUsername(ctx, in.Msg.GetBook(), in.Msg.GetUsername())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&api.GetUserByUsernameResponse{
		User: user,
	}), nil
}

// ListUsers lists users by filters.
func (s *Server) ListUsers(ctx context.Context, in *connect.Request[api.ListUsersRequest]) (*connect.Response[api.ListUsersResponse], error) {
	if err := in.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

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
		fromToken, ok := proto.Clone(p.ListRequest).(*api.ListUsersRequest)
		if !ok {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
		if !proto.Equal(api.StripListUsersPagination(in.Msg), api.StripListUsersPagination(fromToken)) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid page token"))
		}
	}

	users, hasMore, err := s.Repo.ListUsers(ctx, &repo.ListUsersArgs{Book: in.Msg.GetBook(), GreaterThanID: cursor, Users: in.Msg.GetUsers(), Limit: pageSize})
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if hasMore {
		nextPageToken, err = pagination.ToToken(pagination.Pagination{
			Cursor:      users[len(users)-1].GetName(),
			ListRequest: in.Msg,
		})
		if err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&api.ListUsersResponse{
		Users:         users,
		NextPageToken: nextPageToken,
	}), nil
}
