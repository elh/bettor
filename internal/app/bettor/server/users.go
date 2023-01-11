package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

// CreateUser creates a new user.
func (s *Server) CreateUser(context.Context, *connect.Request[api.CreateUserRequest]) (*connect.Response[api.CreateUserResponse], error) {
	return connect.NewResponse(&api.CreateUserResponse{
		User: &api.User{Id: "TODO"},
	}), nil
}
