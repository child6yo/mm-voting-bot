package repository

import (
	"context"
	"time"

	votingbot "github.com/child6yo/mm-voting-bot"
	"github.com/tarantool/go-tarantool/v2"
)

func CreateTarantoolDb(config votingbot.TarantoolConfig) (*tarantool.Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address:  config.TarantoolAddress,
		User:     config.TarantoolUsername,
		Password: config.TarantoolPassword,
	}
	opts := tarantool.Opts{
		Timeout: time.Second,
	}

	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return conn, err
	}
	return conn, nil
}
