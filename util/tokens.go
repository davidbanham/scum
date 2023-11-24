package util

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

func CalcExpiry(days int) string {
	return time.Now().AddDate(0, 0, days).UTC().Format(time.RFC3339)
}

type Token struct {
	expiry time.Time
	token  string
}

func (this Token) String() string {
	return this.token
}

func (this Token) ExpiryString() string {
	return this.expiry.UTC().Format(time.RFC3339)
}

var ErrorTokenExpired = fmt.Errorf("Token expired")
var ErrorTokenInvalid = fmt.Errorf("Token invalid")

func checkTokenExpiry(expiry string) error {
	parsed, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		return errors.New("Invalid expiry string: " + expiry)
	}
	if parsed.Before(time.Now().UTC()) {
		return ErrorTokenExpired
	}
	return nil
}

func CheckToken(secret, expiry, input, token string) error {
	if expiry != "" {
		if err := checkTokenExpiry(expiry); err != nil {
			return err
		}
	}
	plaintext := secret + input + expiry
	hash := sha256.New()
	encHash := hash.Sum([]byte(plaintext))
	target := base64.StdEncoding.EncodeToString(encHash)
	if token != target {
		return ErrorTokenInvalid
	}
	return nil
}

func CalcToken(secret string, days int, input string) Token {
	exp := time.Now().AddDate(0, 0, days).UTC()
	expiry := exp.Format(time.RFC3339)
	if days == 0 {
		expiry = ""
	}
	plaintext := secret + input + expiry
	hash := sha256.New()
	encHash := hash.Sum([]byte(plaintext))
	token := base64.StdEncoding.EncodeToString(encHash)
	return Token{
		expiry: exp,
		token:  token,
	}
}
