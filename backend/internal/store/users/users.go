package users

import (
	"database/sql"
	"fmt"
	"time"

	"fitonex/backend/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Store handles user-related database operations
type Store struct {
	db *sql.DB
}

// New creates a new users store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create creates a new user
func (s *Store) Create(email, password, name string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
		INSERT INTO users (id, email, name, password, created_at, updated_at, premium_until, oauth_provider, oauth_id, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULL, NULL, NULL, NULL)
	`

	_, err = s.db.Exec(query, user.ID, user.Email, user.Name, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *Store) GetByID(id string) (*models.User, error) {
	query := `SELECT id, email, name, password, created_at, updated_at, premium_until, oauth_provider, oauth_id, deleted_at FROM users WHERE id = $1`
	return s.queryUser(query, id)
}

// GetByEmail retrieves a user by email
func (s *Store) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, name, password, created_at, updated_at, premium_until, oauth_provider, oauth_id, deleted_at FROM users WHERE email = $1`
	return s.queryUser(query, email)
}

// Authenticate validates user credentials
func (s *Store) Authenticate(email, password string) (*models.User, error) {
	user, err := s.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

// Update updates a user's information
func (s *Store) Update(id, name, email string) (*models.User, error) {
	query := `
		UPDATE users 
		SET name = $1, email = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, email, name, password, created_at, updated_at, premium_until, oauth_provider, oauth_id, deleted_at
	`

	return s.queryUserRow(query, name, email, time.Now().UTC(), id)
}

func (s *Store) queryUser(query string, args ...any) (*models.User, error) {
	return s.queryUserRow(query, args...)
}

func (s *Store) queryUserRow(query string, args ...any) (*models.User, error) {
	row := s.db.QueryRow(query, args...)
	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.PremiumUntil,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	return user, nil
}

func (s *Store) SetPremium(userID string, until time.Time) error {
	_, err := s.db.Exec(`UPDATE users SET premium_until = $1, updated_at = $2 WHERE id = $3`, until.UTC(), time.Now().UTC(), userID)
	return err
}

func (s *Store) ClearPremium(userID string) error {
	_, err := s.db.Exec(`UPDATE users SET premium_until = NULL, updated_at = $1 WHERE id = $2`, time.Now().UTC(), userID)
	return err
}

func (s *Store) IsPremium(userID string) (bool, error) {
	var premiumUntil sql.NullTime
	err := s.db.QueryRow(`SELECT premium_until FROM users WHERE id = $1`, userID).Scan(&premiumUntil)
	if err != nil {
		return false, err
	}
	if !premiumUntil.Valid {
		return false, nil
	}
	return premiumUntil.Time.After(time.Now()), nil
}

func (s *Store) GetByOAuth(provider, oauthID string) (*models.User, error) {
	query := `SELECT id, email, name, password, created_at, updated_at, premium_until, oauth_provider, oauth_id, deleted_at FROM users WHERE oauth_provider = $1 AND oauth_id = $2`
	return s.queryUser(query, provider, oauthID)
}

func (s *Store) SetOAuth(userID, provider, oauthID string) error {
	_, err := s.db.Exec(`UPDATE users SET oauth_provider = $1, oauth_id = $2, updated_at = $3 WHERE id = $4`, provider, oauthID, time.Now().UTC(), userID)
	return err
}

func (s *Store) UpdatePassword(userID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE users SET password = $1, updated_at = $2 WHERE id = $3`, string(hashedPassword), time.Now().UTC(), userID)
	return err
}

func (s *Store) CreatePasswordResetToken(userID, token string, expiresAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO password_resets (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (token) DO UPDATE SET user_id = EXCLUDED.user_id, expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at
	`, token, userID, expiresAt.UTC(), time.Now().UTC())
	return err
}

func (s *Store) ConsumePasswordResetToken(token string) (string, error) {
	var userID string
	var expiresAt time.Time
	err := s.db.QueryRow(`SELECT user_id, expires_at FROM password_resets WHERE token = $1`, token).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid token")
		}
		return "", err
	}
	if expiresAt.Before(time.Now()) {
		return "", fmt.Errorf("token expired")
	}
	_, _ = s.db.Exec(`DELETE FROM password_resets WHERE token = $1`, token)
	return userID, nil
}

func (s *Store) SoftDelete(userID string) error {
	_, err := s.db.Exec(`UPDATE users SET deleted_at = $1, updated_at = $1 WHERE id = $2`, time.Now().UTC(), userID)
	return err
}
