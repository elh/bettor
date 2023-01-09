package server

import (
	"context"

	api "github.com/elh/bettor/api/bettor/v1alpha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ api.BettorServiceServer = (*Server)(nil)

// Server is an implementation of the Bettor service.
type Server struct {
	api.UnimplementedBettorServiceServer
}

// New initializes a new Server.
func New() *Server {
	return &Server{}
}

// CreateUser creates a new user.
func (s *Server) CreateUser(ctx context.Context, req *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateUser not implemented")
}

// GetUser returns a user by ID.
func (s *Server) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}
