package mem

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/bufbuild/connect-go" // too lazy to isolate errors. repo pkgs will return connect errors
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/entity"
	"github.com/elh/bettor/internal/app/bettor/repo"
	"google.golang.org/protobuf/proto"
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

// hydrate virtual fields like unsettled_centipoints.
func (r *Repo) hydrateUser(ctx context.Context, user *api.User) (*api.User, error) {
	bookID, _ := entity.UserIDs(user.GetName())
	bets, _, err := r.ListBets(ctx, &repo.ListBetsArgs{
		Book:           entity.BookN(bookID),
		User:           user.GetName(),
		ExcludeSettled: true,
		Limit:          1000, // no pagination here
	})
	if err != nil {
		return nil, err
	}

	var unsettledCentipoints uint64
	for _, b := range bets {
		unsettledCentipoints += b.GetCentipoints()
	}
	userCopy := proto.Clone(user).(*api.User)
	userCopy.UnsettledCentipoints = unsettledCentipoints
	return userCopy, nil
}

// CreateUser creates a new user.
func (r *Repo) CreateUser(_ context.Context, user *api.User) error {
	r.userMtx.Lock()
	defer r.userMtx.Unlock()

	user.UnsettledCentipoints = 0 // defensive

	bookID, _ := entity.UserIDs(user.GetName())
	for _, u := range r.Users {
		if u.GetName() == user.GetName() {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("user with id already exists"))
		}
		uBookID, _ := entity.UserIDs(u.GetName())
		if bookID == uBookID && u.Username == user.Username {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("user with username already exists in book"))
		}
	}
	r.Users = append(r.Users, user)
	sort.SliceStable(r.Users, func(i, j int) bool {
		return r.Users[i].Name < r.Users[j].Name
	})
	return nil
}

// UpdateUser updates a user.
func (r *Repo) UpdateUser(_ context.Context, user *api.User) error {
	r.userMtx.Lock()
	defer r.userMtx.Unlock()
	var found bool
	var idx int
	for i, u := range r.Users {
		if u.GetName() == user.GetName() {
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
func (r *Repo) GetUser(ctx context.Context, name string) (*api.User, error) {
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	for _, u := range r.Users {
		if u.GetName() == name {
			u, err := r.hydrateUser(ctx, u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			return u, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
}

// GetUserByUsername gets a user by username.
func (r *Repo) GetUserByUsername(ctx context.Context, book, username string) (*api.User, error) {
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	bookID := entity.BooksIDs(book)
	for _, u := range r.Users {
		uBookID, _ := entity.UserIDs(u.GetName())
		if uBookID == bookID && u.Username == username {
			u, err := r.hydrateUser(ctx, u)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, err)
			}
			return u, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
}

// ListUsers lists users by filters.
func (r *Repo) ListUsers(ctx context.Context, args *repo.ListUsersArgs) (users []*api.User, hasMore bool, err error) {
	r.userMtx.RLock()
	defer r.userMtx.RUnlock()
	bookID := entity.BooksIDs(args.Book)
	var out []*api.User //nolint:prealloc

	var orderedUsers []*api.User
	switch args.OrderBy {
	case "", "name":
		orderedUsers = r.Users
	case "total_centipoints":
		if args.GreaterThanName != "" {
			return nil, false, connect.NewError(connect.CodeInvalidArgument, errors.New("cannot use GreaterThanName with total_centipoints order"))
		}

		var hydratedUsers []*api.User
		for _, u := range r.Users {
			u, err := r.hydrateUser(ctx, u)
			if err != nil {
				return nil, false, connect.NewError(connect.CodeInternal, err)
			}
			hydratedUsers = append(hydratedUsers, u)
		}
		sort.SliceStable(hydratedUsers, func(i, j int) bool {
			return hydratedUsers[i].Centipoints+hydratedUsers[i].UnsettledCentipoints > hydratedUsers[j].Centipoints+hydratedUsers[j].UnsettledCentipoints
		})
		orderedUsers = hydratedUsers
	default:
		return nil, false, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid order by"))
	}

	for _, u := range orderedUsers {
		uBookID, _ := entity.UserIDs(u.GetName())
		if uBookID != bookID {
			continue
		}
		if u.GetName() <= args.GreaterThanName {
			continue
		}
		if len(args.Users) > 0 && !containsStr(args.Users, u.GetName()) {
			continue
		}
		// hydrate
		u, err := r.hydrateUser(ctx, u)
		if err != nil {
			return nil, false, connect.NewError(connect.CodeInternal, errors.New("failed to compute unsettled points"))
		}

		out = append(out, u)
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
func (r *Repo) CreateMarket(_ context.Context, market *api.Market) error {
	r.marketMtx.Lock()
	defer r.marketMtx.Unlock()
	for _, u := range r.Users {
		if u.GetName() == market.GetName() {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("market with id already exists"))
		}
	}
	r.Markets = append(r.Markets, market)
	sort.SliceStable(r.Markets, func(i, j int) bool {
		return r.Markets[i].Name < r.Markets[j].Name
	})
	return nil
}

// UpdateMarket updates a market.
func (r *Repo) UpdateMarket(_ context.Context, market *api.Market) error {
	r.marketMtx.Lock()
	defer r.marketMtx.Unlock()
	var found bool
	var idx int
	for i, m := range r.Markets {
		if m.GetName() == market.GetName() {
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
func (r *Repo) GetMarket(_ context.Context, name string) (*api.Market, error) {
	r.marketMtx.RLock()
	defer r.marketMtx.RUnlock()
	for _, m := range r.Markets {
		if m.GetName() == name {
			return m, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("market not found"))
}

// ListMarkets lists markets by filters.
func (r *Repo) ListMarkets(_ context.Context, args *repo.ListMarketsArgs) (markets []*api.Market, hasMore bool, err error) {
	r.marketMtx.RLock()
	defer r.marketMtx.RUnlock()
	bookID := entity.BooksIDs(args.Book)
	var out []*api.Market //nolint:prealloc
	for _, m := range r.Markets {
		mBookID, _ := entity.MarketIDs(m.GetName())
		if mBookID != bookID {
			continue
		}
		if m.GetName() <= args.GreaterThanName {
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
func (r *Repo) CreateBet(_ context.Context, bet *api.Bet) error {
	r.betMtx.Lock()
	defer r.betMtx.Unlock()
	for _, u := range r.Users {
		if u.GetName() == bet.GetName() {
			return connect.NewError(connect.CodeInvalidArgument, errors.New("bet with id already exists"))
		}
	}
	r.Bets = append(r.Bets, bet)
	sort.SliceStable(r.Bets, func(i, j int) bool {
		return r.Bets[i].Name < r.Bets[j].Name
	})
	return nil
}

// UpdateBet updates a bet.
func (r *Repo) UpdateBet(_ context.Context, bet *api.Bet) error {
	r.betMtx.Lock()
	defer r.betMtx.Unlock()
	var found bool
	var idx int
	for i, b := range r.Bets {
		if b.GetName() == bet.GetName() {
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
func (r *Repo) GetBet(_ context.Context, name string) (*api.Bet, error) {
	r.betMtx.RLock()
	defer r.betMtx.RUnlock()
	for _, b := range r.Bets {
		if b.GetName() == name {
			return b, nil
		}
	}
	return nil, connect.NewError(connect.CodeNotFound, errors.New("bet not found"))
}

// ListBets lists bets by filters.
func (r *Repo) ListBets(_ context.Context, args *repo.ListBetsArgs) (bets []*api.Bet, hasMore bool, err error) {
	r.betMtx.RLock()
	defer r.betMtx.RUnlock()
	bookID := entity.BooksIDs(args.Book)
	var out []*api.Bet //nolint:prealloc
	for _, b := range r.Bets {
		bBookID, _ := entity.BetIDs(b.GetName())
		if bBookID != bookID {
			continue
		}
		if b.GetName() <= args.GreaterThanName {
			continue
		}
		if args.User != "" && b.User != args.User {
			continue
		}
		if args.Market != "" && b.Market != args.Market {
			continue
		}
		if args.ExcludeSettled && b.GetSettledAt() != nil {
			continue
		}
		out = append(out, b)
		if len(out) >= args.Limit+1 {
			break
		}
	}
	if len(out) > args.Limit {
		return out[:args.Limit], true, nil
	}
	return out, false, nil
}

func containsStr(xs []string, y string) bool {
	for _, x := range xs {
		if x == y {
			return true
		}
	}
	return false
}
