package mem

import (
	"context"
	"errors"
	"sync"

	"github.com/bufbuild/connect-go" // too lazy to isolate errors. repo pkgs will return connect errors
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo"
)

var _ repo.Repo = (*Repo)(nil)

// Repo is an in-memory persistence repository.
type Repo struct {
	Users     []*api.User
	Markets   []*api.Market
	Bets      []*api.Bet
	userMtx   sync.RWMutex
	marketMtx sync.RWMutex
	betMtx    sync.RWMutex
}

// CreateUser creates a new user.
func (r *Repo) CreateUser(ctx context.Context, user *api.User) error {
	r.userMtx.Lock()
	defer r.userMtx.Unlock()
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
	r.userMtx.Lock()
	defer r.userMtx.Unlock()
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
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	for _, u := range r.Users {
		if u.Id == id {
			return u, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
}

// GetUserByUsername gets a user by username.
func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*api.User, error) {
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	for _, u := range r.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
}

// ListUsers lists users by filters.
func (r *Repo) ListUsers(ctx context.Context, args *repo.ListUsersArgs) (users []*api.User, hasMore bool, err error) {
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	var out []*api.User
	for _, u := range r.Users {
		if u.Id > args.GreaterThanID {
			out = append(out, u)
		}
		if len(out) >= args.Limit+1 {
			break
		}
	}
	if len(out) > args.Limit {
		return out[:args.Limit], true, nil
	}
	return out, false, nil
}

// CreateMarket creates a new market.
func (r *Repo) CreateMarket(ctx context.Context, market *api.Market) error {
	r.marketMtx.Lock()
	defer r.marketMtx.Unlock()
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
	r.marketMtx.Lock()
	defer r.marketMtx.Unlock()
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
	r.marketMtx.RLock()
	defer r.marketMtx.RUnlock()
	for _, m := range r.Markets {
		if m.Id == id {
			return m, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("market not found"))
}

// ListMarkets lists markets by filters.
func (r *Repo) ListMarkets(ctx context.Context, args *repo.ListMarketsArgs) (markets []*api.Market, hasMore bool, err error) {
	r.marketMtx.RLock()
	defer r.marketMtx.RUnlock()
	var out []*api.Market //nolint:prealloc
	for _, m := range r.Markets {
		if m.Id <= args.GreaterThanID {
			continue
		}
		if args.Status != api.Market_STATUS_UNSPECIFIED && m.Status != args.Status {
			continue
		}
		out = append(out, m)
		if len(out) >= args.Limit+1 {
			break
		}
	}
	if len(out) > args.Limit {
		return out[:args.Limit], true, nil
	}
	return out, false, nil
}

// CreateBet creates a new bet.
func (r *Repo) CreateBet(ctx context.Context, bet *api.Bet) error {
	r.betMtx.Lock()
	defer r.betMtx.Unlock()
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
	r.betMtx.Lock()
	defer r.betMtx.Unlock()
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
	r.betMtx.RLock()
	defer r.betMtx.RUnlock()
	for _, b := range r.Bets {
		if b.Id == id {
			return b, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("bet not found"))
}

// ListBetsByMarket lists bets by market ID.
func (r *Repo) ListBetsByMarket(ctx context.Context, marketID string) ([]*api.Bet, error) {
	r.betMtx.RLock()
	defer r.betMtx.RUnlock()
	var bets []*api.Bet
	for _, b := range r.Bets {
		if b.MarketId == marketID {
			bets = append(bets, b)
		}
	}
	return bets, nil
}
