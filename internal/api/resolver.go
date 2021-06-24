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
	c := graphql.Config{
		Resolvers: &Resolver{
			db: db,
		},
	}

	c.Complexity.Query.Repositories = func(childComplexity int, skip *int, take *int) int {
		takeActual := 10
		if take != nil {
			takeActual = *take
		}
		return childComplexity * takeActual
	}
	c.Complexity.Crd.Spec = func(childComplexity int) int {
		return childComplexity + 1
	}

	return c
}
