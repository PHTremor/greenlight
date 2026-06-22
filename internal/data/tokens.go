package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/PHTremor/greenlight.git/internal/validator"
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

// check the token text provided is 26 bytes long
func ValidateTokenPlainText(v *validator.Validator, tokenPlainText string) {
	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}

// Define the TokenModel type
type TokenModel struct {
	DB *sql.DB
}

// the New() method creates a token and inserts it into the tokens table
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := generateToken(userID, ttl, scope)

	err := m.Insert(token)
	return token, err
}

// Insert() adds a new token into the tokens table
func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// DeleteAlForUser() deletes all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
	DELETE FROM tokens
	WHERE scope=$1 AND user_id=$2`

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
