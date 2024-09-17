package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/BESTSELLER/dependabot-circleci/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type RepoData struct {
	bun.BaseModel `bun:"table:repos"`

	ID       int64        `bun:"id,pk,notnull"`
	Repo     string       `bun:"repo,notnull"`
	Owner    string       `bun:"owner,notnull"`
	Schedule string       `bun:"schedule,notnull"`
	LastRun  bun.NullTime `bun:"lastrun"`
}

// client to be returned instead of creating a new one repeatedly
var client *bun.DB

// According to documentation, it's rarely necessary to close a DB connection.
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connections. Thus, the OpenDB function should be called just once.
func DBClient() *bun.DB {
	if client != nil {
		return client
	}

	dsn := fmt.Sprintf("unix:///cloudsql/%s/.s.PGSQL.5432?sslmode=disable", config.DBConfig.ConnectionName)
	if config.DBConfig.ConnectionString != "" {
		dsn = config.DBConfig.ConnectionString
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDatabase(config.DBConfig.DBName),
		pgdriver.WithUser(config.DBConfig.Username),
		pgdriver.WithPassword(config.DBConfig.Password),
		pgdriver.WithInsecure(true),
		pgdriver.WithDSN(dsn),
	))

	client = bun.NewDB(sqldb, pgdialect.New())
	return client

}

func UpdateRepo(repo RepoData, ctx context.Context) error {
	db := DBClient()
	_, err := db.NewInsert().
		Model(&repo).
		On("CONFLICT (id) DO UPDATE").
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func GetRepos(ctx context.Context) (repos []RepoData, err error) {
	db := DBClient()
	err = db.NewSelect().Table("repos").Scan(ctx, &repos)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func DeleteRepo(repo RepoData, ctx context.Context) error {
	db := DBClient()
	_, err := db.NewDelete().
		Model(&repo).
		Where("id = ?", repo.ID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
