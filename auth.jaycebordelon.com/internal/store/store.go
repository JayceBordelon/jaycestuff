package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

type User struct {
	ID            int64
	GoogleSub     string
	Email         string
	EmailVerified bool
	Name          string
	PictureURL    string
	CreatedAt     time.Time
	LastLoginAt   time.Time
}

type Session struct {
	ID         int64
	UserID     int64
	ClientID   sql.NullString
	CreatedAt  time.Time
	LastUsedAt time.Time
	ExpiresAt  time.Time
	User       User
}

type OAuthClient struct {
	ClientID         string
	Name             string
	RedirectURIs     []string
	ClientSecretHash []byte
}

type AuthCode struct {
	CodeHash    []byte
	UserID      int64
	ClientID    string
	RedirectURI string
	ExpiresAt   time.Time
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Ping() error  { return s.db.Ping() }
func (s *Store) DB() *sql.DB  { return s.db }

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			google_sub TEXT UNIQUE NOT NULL,
			email TEXT NOT NULL,
			email_verified BOOLEAN NOT NULL DEFAULT false,
			name TEXT NOT NULL DEFAULT '',
			picture_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

		CREATE TABLE IF NOT EXISTS oauth_states (
			state TEXT PRIMARY KEY,
			client_id TEXT,
			redirect_uri TEXT NOT NULL DEFAULT '',
			consumer_state TEXT NOT NULL DEFAULT '',
			return_to TEXT NOT NULL DEFAULT '/',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS oauth_clients (
			client_id TEXT PRIMARY KEY,
			client_secret_hash BYTEA NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			redirect_uris TEXT[] NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth_codes (
			code_hash BYTEA PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			client_id TEXT NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
			redirect_uri TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			client_id TEXT REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
			token_hash BYTEA UNIQUE NOT NULL,
			user_agent TEXT NOT NULL DEFAULT '',
			ip_addr INET,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	`)
	return err
}

// --- Users ---

func (s *Store) UpsertUser(googleSub, email string, emailVerified bool, name, pictureURL string) (int64, error) {
	var id int64
	err := s.db.QueryRow(`
		INSERT INTO users (google_sub, email, email_verified, name, picture_url, last_login_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (google_sub) DO UPDATE SET
			email = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified,
			name = EXCLUDED.name,
			picture_url = EXCLUDED.picture_url,
			last_login_at = NOW()
		RETURNING id
	`, googleSub, email, emailVerified, name, pictureURL).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upsert user: %w", err)
	}
	return id, nil
}

func (s *Store) GetUser(userID int64) (*User, error) {
	var u User
	err := s.db.QueryRow(`
		SELECT id, google_sub, email, email_verified, name, picture_url, created_at, last_login_at
		FROM users WHERE id = $1
	`, userID).Scan(&u.ID, &u.GoogleSub, &u.Email, &u.EmailVerified, &u.Name, &u.PictureURL, &u.CreatedAt, &u.LastLoginAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

// --- oauth_states (Google CSRF + authorize stash) ---

func (s *Store) CreateOAuthState(state, clientID, redirectURI, consumerState, returnTo string, ttl time.Duration) error {
	var clientArg any
	if clientID != "" {
		clientArg = clientID
	}
	_, err := s.db.Exec(`
		INSERT INTO oauth_states (state, client_id, redirect_uri, consumer_state, return_to, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW() + ($6 || ' seconds')::interval)
	`, state, clientArg, redirectURI, consumerState, returnTo, fmt.Sprintf("%d", int64(ttl.Seconds())))
	if err != nil {
		return fmt.Errorf("create oauth state: %w", err)
	}
	return nil
}

type OAuthStateRow struct {
	ClientID      string
	RedirectURI   string
	ConsumerState string
	ReturnTo      string
}

func (s *Store) ConsumeOAuthState(state string) (*OAuthStateRow, error) {
	var row OAuthStateRow
	var clientID sql.NullString
	err := s.db.QueryRow(`
		DELETE FROM oauth_states
		WHERE state = $1 AND expires_at > NOW()
		RETURNING client_id, redirect_uri, consumer_state, return_to
	`, state).Scan(&clientID, &row.RedirectURI, &row.ConsumerState, &row.ReturnTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oauth state not found or expired")
		}
		return nil, fmt.Errorf("consume oauth state: %w", err)
	}
	if clientID.Valid {
		row.ClientID = clientID.String
	}
	return &row, nil
}

// --- oauth_clients ---

func (s *Store) UpsertClient(clientID string, secretHash []byte, name string, redirectURIs []string) error {
	_, err := s.db.Exec(`
		INSERT INTO oauth_clients (client_id, client_secret_hash, name, redirect_uris)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (client_id) DO UPDATE SET
			client_secret_hash = EXCLUDED.client_secret_hash,
			name = EXCLUDED.name,
			redirect_uris = EXCLUDED.redirect_uris
	`, clientID, secretHash, name, pq.Array(redirectURIs))
	if err != nil {
		return fmt.Errorf("upsert client: %w", err)
	}
	return nil
}

func (s *Store) GetClient(clientID string) (*OAuthClient, error) {
	var c OAuthClient
	var uris []string
	err := s.db.QueryRow(`
		SELECT client_id, client_secret_hash, name, redirect_uris
		FROM oauth_clients WHERE client_id = $1
	`, clientID).Scan(&c.ClientID, &c.ClientSecretHash, &c.Name, pq.Array(&uris))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get client: %w", err)
	}
	c.RedirectURIs = uris
	return &c, nil
}

// --- auth_codes ---

func (s *Store) CreateAuthCode(codeHash []byte, userID int64, clientID, redirectURI string, ttl time.Duration) error {
	_, err := s.db.Exec(`
		INSERT INTO auth_codes (code_hash, user_id, client_id, redirect_uri, expires_at)
		VALUES ($1, $2, $3, $4, NOW() + ($5 || ' seconds')::interval)
	`, codeHash, userID, clientID, redirectURI, fmt.Sprintf("%d", int64(ttl.Seconds())))
	if err != nil {
		return fmt.Errorf("create auth code: %w", err)
	}
	return nil
}

// ConsumeAuthCode deletes and returns the code's details. One-shot.
func (s *Store) ConsumeAuthCode(codeHash []byte) (*AuthCode, error) {
	var ac AuthCode
	err := s.db.QueryRow(`
		DELETE FROM auth_codes
		WHERE code_hash = $1 AND expires_at > NOW()
		RETURNING code_hash, user_id, client_id, redirect_uri, expires_at
	`, codeHash).Scan(&ac.CodeHash, &ac.UserID, &ac.ClientID, &ac.RedirectURI, &ac.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("auth code not found or expired")
		}
		return nil, fmt.Errorf("consume auth code: %w", err)
	}
	return &ac, nil
}

// --- sessions ---

func (s *Store) CreateSession(userID int64, clientID string, tokenHash []byte, userAgent, ipAddr string, ttl time.Duration) error {
	var ipArg any
	if ipAddr != "" {
		ipArg = ipAddr
	}
	var clientArg any
	if clientID != "" {
		clientArg = clientID
	}
	_, err := s.db.Exec(`
		INSERT INTO sessions (user_id, client_id, token_hash, user_agent, ip_addr, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW() + ($6 || ' seconds')::interval)
	`, userID, clientArg, tokenHash, userAgent, ipArg, fmt.Sprintf("%d", int64(ttl.Seconds())))
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *Store) LookupSession(tokenHash []byte) (*Session, error) {
	var sess Session
	err := s.db.QueryRow(`
		SELECT s.id, s.user_id, s.client_id, s.created_at, s.last_used_at, s.expires_at,
			u.id, u.google_sub, u.email, u.email_verified, u.name, u.picture_url, u.created_at, u.last_login_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
			AND s.revoked_at IS NULL
			AND s.expires_at > NOW()
	`, tokenHash).Scan(
		&sess.ID, &sess.UserID, &sess.ClientID, &sess.CreatedAt, &sess.LastUsedAt, &sess.ExpiresAt,
		&sess.User.ID, &sess.User.GoogleSub, &sess.User.Email, &sess.User.EmailVerified,
		&sess.User.Name, &sess.User.PictureURL, &sess.User.CreatedAt, &sess.User.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup session: %w", err)
	}
	return &sess, nil
}

func (s *Store) TouchSession(sessionID int64) error {
	_, err := s.db.Exec(`
		UPDATE sessions SET last_used_at = NOW()
		WHERE id = $1 AND last_used_at < NOW() - INTERVAL '1 hour'
	`, sessionID)
	if err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (s *Store) RevokeSession(sessionID int64) error {
	_, err := s.db.Exec(`UPDATE sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *Store) RevokeAllSessionsForUser(userID int64) error {
	_, err := s.db.Exec(`UPDATE sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("revoke user sessions: %w", err)
	}
	return nil
}

func (s *Store) SweepExpired() (sessions, states, codes int64, err error) {
	res, err := s.db.Exec(`
		DELETE FROM sessions
		WHERE expires_at < NOW() - INTERVAL '7 days'
			OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days')
	`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("sweep sessions: %w", err)
	}
	sessions, _ = res.RowsAffected()

	res, err = s.db.Exec(`DELETE FROM oauth_states WHERE expires_at < NOW()`)
	if err != nil {
		return sessions, 0, 0, fmt.Errorf("sweep oauth_states: %w", err)
	}
	states, _ = res.RowsAffected()

	res, err = s.db.Exec(`DELETE FROM auth_codes WHERE expires_at < NOW()`)
	if err != nil {
		return sessions, states, 0, fmt.Errorf("sweep auth_codes: %w", err)
	}
	codes, _ = res.RowsAffected()

	return sessions, states, codes, nil
}
