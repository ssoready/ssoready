package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type Store struct {
	q *queries.Queries
}

func New(db *pgxpool.Pool) *Store {
	return &Store{q: queries.New(db)}
}
