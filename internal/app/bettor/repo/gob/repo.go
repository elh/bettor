package gob

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/bufbuild/connect-go" // too lazy to isolate errors :shrug:
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
)

var _ repo.Repo = (*Repo)(nil)

func init() {
	// interface types need to be registered
	gob.Register(&api.Market_Pool{})
}

// Repo is an file-backed gob persistence repository.
type Repo struct {
	Mem      *mem.Repo
	FileName string
}

// New initializes a Gob-backed repository.
func New(fileName string) (*Repo, error) {
	file, err := os.Open(fileName)
	if errors.Is(err, fs.ErrNotExist) {
		if _, err := os.Create(fileName); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("gob file could not be created: %w", err))
		}
		r := &Repo{
			Mem:      &mem.Repo{},
			FileName: fileName,
		}
		return r, r.persist()
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("gob file could not be opened: %w", err))
	}
	defer file.Close()
	m := mem.Repo{}
	if err := gob.NewDecoder(file).Decode(&m); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("gob file could not be decoded: %w", err))
	}

	return &Repo{
		Mem:      &m,
		FileName: fileName,
	}, nil
}

func (r *Repo) persist() error {
	file, err := os.OpenFile(r.FileName, os.O_WRONLY, os.ModeExclusive)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("gob file could not be opened: %w", err))
	}
	defer file.Close()
	if err := gob.NewEncoder(file).Encode(r.Mem); err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("gob file could not be encoded: %w", err))
	}
	return nil
}

// CreateUser creates a new user.
func (r *Repo) CreateUser(ctx context.Context, user *api.User) error {
	if err := r.Mem.CreateUser(ctx, user); err != nil {
		return err
	}
	return r.persist()
}

// UpdateUser updates a user.
func (r *Repo) UpdateUser(ctx context.Context, user *api.User) error {
	if err := r.Mem.UpdateUser(ctx, user); err != nil {
		return err
	}
	return r.persist()
}

// GetUser gets a user by ID.
func (r *Repo) GetUser(ctx context.Context, id string) (*api.User, error) {
	return r.Mem.GetUser(ctx, id)
}

// CreateMarket creates a new market.
func (r *Repo) CreateMarket(ctx context.Context, market *api.Market) error {
	if err := r.Mem.CreateMarket(ctx, market); err != nil {
		return err
	}
	return r.persist()
}

// UpdateMarket updates a market.
func (r *Repo) UpdateMarket(ctx context.Context, market *api.Market) error {
	if err := r.Mem.UpdateMarket(ctx, market); err != nil {
		return err
	}
	return r.persist()
}

// GetMarket gets a market by ID.
func (r *Repo) GetMarket(ctx context.Context, id string) (*api.Market, error) {
	return r.Mem.GetMarket(ctx, id)
}

// CreateBet creates a new user.
func (r *Repo) CreateBet(ctx context.Context, bet *api.Bet) error {
	if err := r.Mem.CreateBet(ctx, bet); err != nil {
		return err
	}
	return r.persist()
}

// UpdateBet updates a bet.
func (r *Repo) UpdateBet(ctx context.Context, bet *api.Bet) error {
	if err := r.Mem.UpdateBet(ctx, bet); err != nil {
		return err
	}
	return r.persist()
}

// GetBet gets a bet by ID.
func (r *Repo) GetBet(ctx context.Context, id string) (*api.Bet, error) {
	return r.Mem.GetBet(ctx, id)
}

// ListBetsByMarket lists bets by market ID.
func (r *Repo) ListBetsByMarket(ctx context.Context, marketID string) ([]*api.Bet, error) {
	return r.Mem.ListBetsByMarket(ctx, marketID)
}
