package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ssoready/ssoready/internal/pagetoken"
	"github.com/ssoready/ssoready/internal/statesign"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Store struct {
	db                   *pgxpool.Pool
	q                    *queries.Queries
	pageEncoder          pagetoken.Encoder
	globalDefaultAuthURL string
	statesigner          *statesign.Signer
}

type NewStoreParams struct {
	DB                   *pgxpool.Pool
	PageEncoder          pagetoken.Encoder
	GlobalDefaultAuthURL string
	SAMLStateSigningKey  [32]byte
}

func New(p NewStoreParams) *Store {
	return &Store{
		db:                   p.DB,
		q:                    queries.New(p.DB),
		pageEncoder:          p.PageEncoder,
		globalDefaultAuthURL: p.GlobalDefaultAuthURL,
		statesigner:          &statesign.Signer{Key: p.SAMLStateSigningKey},
	}
}

func (s *Store) tx(ctx context.Context) (tx pgx.Tx, q *queries.Queries, commit func() error, rollback func() error, err error) {
	tx, err = s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("begin tx: %w", err)
	}

	commit = func() error { return tx.Commit(ctx) }
	rollback = func() error { return tx.Rollback(ctx) }
	return tx, queries.New(tx), commit, rollback, nil
}

func derefOrEmpty[T any](t *T) T {
	var z T
	if t == nil {
		return z
	}
	return *t
}

func ptrTimeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
