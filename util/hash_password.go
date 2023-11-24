package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(rawpassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(rawpassword), 10)
	return string(hash), err
}
