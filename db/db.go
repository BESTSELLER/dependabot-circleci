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

func DBClient() *bun.DB {
	dsn := fmt.Sprintf("unix://%s:%s@/%s/cloudsql/%s/.s.PGSQL.5432", config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.DBName, config.DBConfig.ConnectionName)
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	return bun.NewDB(sqldb, pgdialect.New())

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
