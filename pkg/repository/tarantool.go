package repository

import (
	"context"
	"time"

	"github.com/tarantool/go-tarantool/v2"
)

func CreateTarantoolDb() (*tarantool.Connection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	dialer := tarantool.NetDialer{
		Address:  "0.0.0.0:3301",
		User:     "votingbot",
		Password: "123456",
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
