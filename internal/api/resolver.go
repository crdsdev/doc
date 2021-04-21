package api

import (
	graphql "github.com/crdsdev/doc/internal/api/graphql"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) graphql.Config {
	return graphql.Config{
		Resolvers: &Resolver{
			db: db,
		},
	}
}
