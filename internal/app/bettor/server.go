package server

import api "github.com/elh/bettor/gen/bettor/v1alpha"

var _ api.BettorServiceServer = (*Server)(nil)

// Server is an implementation of the Bettor service.
type Server struct {
	api.UnimplementedBettorServiceServer
}
