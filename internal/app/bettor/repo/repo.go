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
	GetUserByUsername(ctx context.Context, username string) (*api.User, error)
	ListUsers(ctx context.Context, args *ListUsersArgs) (users []*api.User, hasMore bool, err error)
	CreateMarket(ctx context.Context, market *api.Market) error
	UpdateMarket(ctx context.Context, market *api.Market) error
	GetMarket(ctx context.Context, id string) (*api.Market, error)
	ListMarkets(ctx context.Context, args *ListMarketsArgs) (markets []*api.Market, hasMore bool, err error)
	CreateBet(ctx context.Context, bet *api.Bet) error
	UpdateBet(ctx context.Context, bet *api.Bet) error
	GetBet(ctx context.Context, id string) (*api.Bet, error)
	ListBets(ctx context.Context, args *ListBetsArgs) (bets []*api.Bet, hasMore bool, err error)
}

// ListUsersArgs are the arguments for listing users.
type ListUsersArgs struct {
	GreaterThanID string
	Users         []string
	Limit         int
}

// ListMarketsArgs are the arguments for listing markets.
type ListMarketsArgs struct {
	GreaterThanID string
	Status        api.Market_Status
	Limit         int
}

// ListBetsArgs are the arguments for listing bets.
type ListBetsArgs struct {
	GreaterThanID  string
	User           string
	Market         string
	ExcludeSettled bool
	Limit          int
}
