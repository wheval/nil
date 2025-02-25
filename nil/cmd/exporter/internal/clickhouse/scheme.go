package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/NilFoundation/nil/nil/common/check"
)

var tableSchemeCache map[string]reflectedScheme = nil

func initSchemeCache() map[string]reflectedScheme {
	tableScheme := make(map[string]reflectedScheme)

	blockScheme, err := reflectSchemeToClickhouse(&BlockWithBinary{})
	check.PanicIfErr(err)

	tableScheme["blocks"] = blockScheme
	transactionScheme, err := reflectSchemeToClickhouse(&TransactionWithBinary{})
	check.PanicIfErr(err)

	tableScheme["transactions"] = transactionScheme
	logScheme, err := reflectSchemeToClickhouse(&LogWithBinary{})
	check.PanicIfErr(err)

	tableScheme["logs"] = logScheme

	return tableScheme
}

func getTableScheme() map[string]reflectedScheme {
	if tableSchemeCache == nil {
		tableSchemeCache = initSchemeCache()
	}

	return tableSchemeCache
}

func setupScheme(ctx context.Context, conn driver.Conn, tableName string, keys []string) error {
	tableScheme := getTableScheme()
	scheme, ok := tableScheme[tableName]
	if !ok {
		return fmt.Errorf("scheme for %s not found", tableName)
	}

	if err := conn.Exec(ctx, scheme.CreateTableQuery(
		tableName,
		"ReplacingMergeTree",
		keys,
		keys,
	)); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}

func setupSchemes(ctx context.Context, conn driver.Conn) error {
	if err := setupScheme(ctx, conn,
		"blocks", []string{"shard_id", "hash"}); err != nil {
		return err
	}

	if err := setupScheme(ctx, conn,
		"transactions", []string{"hash", "outgoing"}); err != nil {
		return err
	}

	if err := setupScheme(ctx, conn,
		"logs", []string{"transaction_hash"}); err != nil {
		return err
	}

	return nil
}

func createTableQuery(tableName, fields, engine string, primaryKeys, orderKeys []string) string {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s
		(%s)
		ENGINE = %s
		PRIMARY KEY (%s)
		order by (%s)
`, tableName, fields, engine, strings.Join(primaryKeys, ", "), strings.Join(orderKeys, ", "))
	logger.Debug().Msgf("CreateTableQuery: %s", query)
	return query
}
