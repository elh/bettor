package repo

import (
	"context"

	api "github.com/elh/bettor/api/bettor/v1alpha"
)

// Repo is a persistence repository.
type Repo interface {
	CreateUser(ctx context.Context, user *api.User) error
	GetUser(ctx context.Context, id string) (*api.User, error)
	CreateMarket(ctx context.Context, market *api.Market) error
}
