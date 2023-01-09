package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
)

var _ bettorv1alphaconnect.BettorServiceHandler = (*Server)(nil)

// Server is an implementation of the Bettor service.
type Server struct {
	bettorv1alphaconnect.UnimplementedBettorServiceHandler
}

// New initializes a new Server.
func New() *Server {
	return &Server{}
}

// CreateUser creates a new user.
func (s *Server) CreateUser(context.Context, *connect.Request[api.CreateUserRequest]) (*connect.Response[api.CreateUserResponse], error) {
	return connect.NewResponse(&api.CreateUserResponse{
		User: &api.User{Id: "TODO"},
	}), nil
}
