package server

import (
	"sync"

	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/repo"
	"github.com/go-kit/log"
)

var _ bettorv1alphaconnect.BettorServiceHandler = (*Server)(nil)

// Server is an implementation of the Bettor service.
type Server struct {
	bettorv1alphaconnect.UnimplementedBettorServiceHandler
	Repo      repo.Repo
	Logger    log.Logger
	marketMtx sync.RWMutex // this is conservative and inelegant
}

// New initializes a new Server.
func New(repo repo.Repo, logger log.Logger) *Server {
	return &Server{
		Repo:   repo,
		Logger: logger,
	}
}
