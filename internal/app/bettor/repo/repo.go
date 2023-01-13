package repo

import (
	"context"

	api "github.com/elh/bettor/api/bettor/v1alpha"
)

// Repo is a persistence repository.
type Repo interface {
	CreateUser(ctx context.Context, user *api.User) error
	UpdateUser(ctx context.Context, user *api.User) error
	GetUser(ctx context.Context, id string) (*api.User, error)
	CreateMarket(ctx context.Context, market *api.Market) error
	UpdateMarket(ctx context.Context, market *api.Market) error
	GetMarket(ctx context.Context, id string) (*api.Market, error)
	CreateBet(ctx context.Context, bet *api.Bet) error
	UpdateBet(ctx context.Context, bet *api.Bet) error
	GetBet(ctx context.Context, id string) (*api.Bet, error)
	ListBetsByMarket(ctx context.Context, marketID string) ([]*api.Bet, error) // TODO: generalize and paginate
}
