package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/PHTremor/greenlight.git/internal/validator"
)

// custom errDuplicateEmail error
var (
	errDuplicateEmail = errors.New("duplicate email")
)

// Password & Version use json:"-" to prevent the fields appearing in the output
// when we encode to json
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int64     `json:"-"`
}

type password struct {
	Plaintext *string
	hash      []byte
}

// Set() calculates the bcrypt hash of a plaintext password and
// stores both the plaintext and hash in the struct
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.Plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches() checks whether the given plaintext password matches the hashed password stored in the struct
// retuns true if matched or false otherwise
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlainText(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be 500 bytes long")

	// validate the email
	ValidateEmail(v, user.Email)

	// validate password if not nil
	if user.Password.Plaintext != nil {
		ValidatePasswordPlainText(v, *user.Password.Plaintext)
	}

	// a nil hashpassword would be caused by a logic error, it shouldnt happen
	// so we'll panic
	if user.Password.hash == nil {
		panic("missing hash password for user")
	}
}

// Wrap the commection pool in a UserModel
type UserModel struct {
	DB *sql.DB
}

// Insert a new user record into the database
func (m UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (name. email, password_harsh, activated)
	VALUES ($1,$2,$3,$4)
	RETURNING id, created_at, version`

	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// if user with the email exist, an insert will violate the UNIQUE
	// "users_email_key" constraint we set up
	// check for this error and return the custom ErrDuplicateEmail message
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return errDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Retrieve user details from the database based on the user's email address
func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT is, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1`

	var user User

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// update the deatils of a specific user
func (m UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE is = $5 AND version = $6
	RETURNING version`

	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return errDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
