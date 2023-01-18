package server

import (
	"fmt"
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
func New(args ...Arg) (*Server, error) {
	serverArgs := &serverArgs{
		logger: log.NewNopLogger(),
	}
	for _, arg := range args {
		arg(serverArgs)
	}
	if serverArgs.repo == nil || serverArgs.logger == nil {
		return nil, fmt.Errorf("missing required arguments")
	}

	return &Server{
		Repo:   serverArgs.repo,
		Logger: serverArgs.logger,
	}, nil
}

type serverArgs struct {
	repo   repo.Repo
	logger log.Logger
}

// Arg is an argument for constructing a Server.
type Arg func(o *serverArgs)

// WithRepo provides a repo to the Server.
func WithRepo(repo repo.Repo) Arg {
	return Arg(func(a *serverArgs) {
		a.repo = repo
	})
}

// WithLogger provides a logger to the Server.
func WithLogger(logger log.Logger) Arg {
	return Arg(func(a *serverArgs) {
		a.logger = logger
	})
}
