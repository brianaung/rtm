package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

type user struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

func (s *service) addUser(ctx context.Context, u *user) (*user, error) {
	u.ID = uuid.Must(uuid.NewV4())
	_, err := s.db.Exec(ctx, `insert into "user"(id, username, email, password) values($1, $2, $3, $4)`, u.ID, u.Username, u.Email, u.Password)
	if err != nil {
		return nil, err
	}
	return u, nil
}
