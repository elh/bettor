package server

import (
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/repo"
)

var _ bettorv1alphaconnect.BettorServiceHandler = (*Server)(nil)

// Server is an implementation of the Bettor service.
type Server struct {
	bettorv1alphaconnect.UnimplementedBettorServiceHandler
	Repo repo.Repo // NOTE: no tx right now
}

// New initializes a new Server.
func New(repo repo.Repo) *Server {
	return &Server{
		Repo: repo,
	}
}
