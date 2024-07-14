package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ucarion/cli"
)

//go:embed migrations
var migrateFS embed.FS

func main() {
	cli.Run(context.Background(), version, force, up)
}

type args struct {
	Database string `cli:"-d,--database"`
	Verbose  bool   `cli:"-v,--verbose"`
}

func (a args) migrate() (*migrate.Migrate, error) {
	src, err := iofs.New(migrateFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("create migrate source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, a.Database)
	if err != nil {
		return nil, fmt.Errorf("create migrate: %w", err)
	}

	m.Log = logger{verbose: a.Verbose}
	return m, nil
}

type logger struct {
	verbose bool
}

func (l logger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l logger) Verbose() bool {
	return l.verbose
}

type versionArgs struct {
	Args args `cli:"version,subcmd"`
}

func version(_ context.Context, args versionArgs) error {
	m, err := args.Args.migrate()
	if err != nil {
		return err
	}

	v, dirty, err := m.Version()
	if err != nil {
		return err
	}

	if dirty {
		fmt.Printf("%d (dirty)\n", v)
	} else {
		fmt.Printf("%d\n", v)
	}
	return nil
}

type forceArgs struct {
	Args    args `cli:"force,subcmd"`
	Version int  `cli:"version"`
}

func force(_ context.Context, args forceArgs) error {
	m, err := args.Args.migrate()
	if err != nil {
		return err
	}

	if err := m.Force(args.Version); err != nil {
		return err
	}
	return nil
}

type upArgs struct {
	Args args `cli:"up,subcmd"`
}

func up(_ context.Context, args upArgs) error {
	m, err := args.Args.migrate()
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}
	return nil
}
