package user

import "golang.org/x/crypto/bcrypt"

func hashAndSalt(password string) (string, error) {
	// GenerateFromPassword salt the password for us aside from hashing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashedPassword), err
}
