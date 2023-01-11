package server

import (
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
