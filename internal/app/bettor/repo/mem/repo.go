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

// GetMarket gets a market by ID.
func (r *Repo) GetMarket(ctx context.Context, id string) (*api.Market, error) {
	for _, m := range r.Markets {
		if m.Id == id {
			return m, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("market not found"))
}
