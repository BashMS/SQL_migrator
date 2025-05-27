package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/BashMS/SQL_migrator/pkg/config" //nolint:depguard
	"github.com/jackc/pgx/v4"                   //nolint:depguard
)

const configPath = "/src/.bin/config.yml"

func CloseConnectDB(conn *pgx.Conn) {
	if conn == nil {
		return
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	_ = conn.Close(ctx)
}

func ConnectDB() (*pgx.Conn, error) {
	config := config.Config{}
	filePath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	if err := config.ReadConfigFromFile(filePath); err != nil {
		return nil, err
	}
	config.Apply()
	if config.DSN == "" {
		return nil, fmt.Errorf("empty string to connect")
	}
	connCtx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()

	connConfig, err := pgx.ParseConfig(config.DSN)
	if err != nil {
		return nil, err
	}
	conn, err := pgx.ConnectConfig(connCtx, connConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
