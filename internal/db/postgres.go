package db

// Minimal Postgres client abstraction for MVP. Replaces direct DB calls with a pluggable store later.
type Config struct {
  DSN string
}

type Client struct{}

// Connect returns a new Postgres client. In MVP this is a no-op facade.
func Connect(cfg Config) (*Client, error) {
  return &Client{}, nil
}

// Close closes the database connection if applicable (no-op for MVP).
func (c *Client) Close() error { return nil }
