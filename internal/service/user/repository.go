package user

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

func addUser(ctx context.Context, db *pgxpool.Pool, u *User) (*User, error) {
	u.ID = uuid.Must(uuid.NewV4())
	_, err := db.Exec(ctx, `insert into "user"(id, username, email, password) values($1, $2, $3, $4)`, u.ID, u.Username, u.Email, u.Password)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func getUserByName(ctx context.Context, db *pgxpool.Pool, username string) (*User, error) {
	u := &User{}
	err := db.QueryRow(ctx, `select * from "user" where username = $1`, username).Scan(&u.ID, &u.Username, &u.Email, &u.Password)
	if err != nil {
		return nil, err
	}
	return u, nil
}
