package mem

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go" // too lazy to isolate errors :shrug:
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/repo"
)

var _ repo.Repo = (*Repo)(nil)

// Repo is an in-memory persistence repository.
type Repo struct {
	Users []*api.User
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
