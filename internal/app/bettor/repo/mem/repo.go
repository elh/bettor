package mem

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go" // too lazy to isolate errors. repo pkgs will return connect errors
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo"
)

var _ repo.Repo = (*Repo)(nil)

// Repo is an in-memory persistence repository.
type Repo struct {
	Users   []*api.User
	Markets []*api.Market
	Bets    []*api.Bet
}

// CreateUser creates a new user.
func (r *Repo) CreateUser(ctx context.Context, user *api.User) error {
	for _, u := range r.Users {
		if u.Id == user.Id {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("user with id already exists"))
		}
		if u.Username == user.Username {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("user with username already exists"))
		}
	}
	r.Users = append(r.Users, user)
	return nil
}

// UpdateUser updates a user.
func (r *Repo) UpdateUser(ctx context.Context, user *api.User) error {
	var found bool
	var idx int
	for i, u := range r.Users {
		if u.Id == user.Id {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	}
	r.Users[idx] = user
	return nil
}

// GetUser gets a user by ID.
func (r *Repo) GetUser(ctx context.Context, id string) (*api.User, error) {
	for _, u := range r.Users {
		if u.Id == id {
			return u, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
}

// CreateMarket creates a new market.
func (r *Repo) CreateMarket(ctx context.Context, market *api.Market) error {
	for _, u := range r.Users {
		if u.Id == market.Id {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("market with id already exists"))
		}
	}
	r.Markets = append(r.Markets, market)
	return nil
}

// UpdateMarket updates a market.
func (r *Repo) UpdateMarket(ctx context.Context, market *api.Market) error {
	var found bool
	var idx int
	for i, m := range r.Markets {
		if m.Id == market.Id {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return connect.NewError(connect.CodeNotFound, errors.New("market not found"))
	}
	r.Markets[idx] = market
	return nil
}

// GetMarket gets a market by ID.
func (r *Repo) GetMarket(ctx context.Context, id string) (*api.Market, error) {
	for _, m := range r.Markets {
		if m.Id == id {
			return m, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("market not found"))
}

// CreateBet creates a new bet.
func (r *Repo) CreateBet(ctx context.Context, bet *api.Bet) error {
	for _, u := range r.Users {
		if u.Id == bet.Id {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("bet with id already exists"))
		}
	}
	r.Bets = append(r.Bets, bet)
	return nil
}

// UpdateBet updates a bet.
func (r *Repo) UpdateBet(ctx context.Context, bet *api.Bet) error {
	var found bool
	var idx int
	for i, b := range r.Bets {
		if b.Id == bet.Id {
			found = true
			idx = i
			break
		}
	}
	if !found {
		return connect.NewError(connect.CodeNotFound, errors.New("bet not found"))
	}
	r.Bets[idx] = bet
	return nil
}

// GetBet gets a bet by ID.
func (r *Repo) GetBet(ctx context.Context, id string) (*api.Bet, error) {
	for _, b := range r.Bets {
		if b.Id == id {
			return b, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("bet not found"))
}

// ListBetsByMarket lists bets by market ID.
func (r *Repo) ListBetsByMarket(ctx context.Context, marketID string) ([]*api.Bet, error) {
	var bets []*api.Bet
	for _, b := range r.Bets {
		if b.MarketId == marketID {
			bets = append(bets, b)
		}
	}
	return bets, nil
}
