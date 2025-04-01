package common

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func CreateClickHouseDbIfNotExists(ctx context.Context, dbname, user, password, endpoint string) error {
	connectionOptions := clickhouse.Options{
		Auth: clickhouse.Auth{
			Database: "system",
			Username: user,
			Password: password,
		},
		Addr: []string{endpoint},
	}
	conn, err := clickhouse.Open(&connectionOptions)
	if err != nil {
		return err
	}
	defer conn.Close()

	rows, err := conn.Query(ctx, "SELECT name FROM system.databases WHERE name = ?", dbname)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbname))
		return err
	}
	return nil
}
