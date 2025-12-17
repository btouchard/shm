// SPDX-License-Identifier: AGPL-3.0-or-later

// Package postgres provides PostgreSQL implementations of the repository interfaces.
package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Store holds the database connection and provides access to repositories.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store with a database connection.
func NewStore(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying database connection for advanced use cases.
func (s *Store) DB() *sql.DB {
	return s.db
}

// InstanceRepository returns an InstanceRepository backed by this store.
func (s *Store) InstanceRepository() *InstanceRepository {
	return NewInstanceRepository(s.db)
}

// SnapshotRepository returns a SnapshotRepository backed by this store.
func (s *Store) SnapshotRepository() *SnapshotRepository {
	return NewSnapshotRepository(s.db)
}

// DashboardReader returns a DashboardReader backed by this store.
func (s *Store) DashboardReader() *DashboardReader {
	return NewDashboardReader(s.db)
}
