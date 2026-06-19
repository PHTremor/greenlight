package data

import (
	"crypto/rand"
	"crypto/sha256"
	"time"
)

// define constants for the token scope
const (
	ScopeActivation = "Activation"
)

// a Token struct to hold data for an individual token
type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) *Token {
	token := &Token{
		Plaintext: rand.Text(), // generate a random text/token
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	// generate a SHA256 hash of the token string
	hash := sha256.Sum256([]byte(token.Plaintext))
	// sha256.Sum256() return an array of length 32, we'll convert it to a slice to easily work with it
	token.Hash = hash[:]

	return token
}
