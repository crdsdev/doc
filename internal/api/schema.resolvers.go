package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	graphql1 "github.com/crdsdev/doc/internal/api/graphql"
	"github.com/crdsdev/doc/internal/api/models"
	"github.com/jackc/pgtype"
)

func (r *crdResolver) Tag(ctx context.Context, obj *models.Crd) (*models.Tag, error) {
	row := r.db.QueryRow(ctx, "SELECT DISTINCT name, time FROM tags WHERE id = $1 LIMIT 1", obj.TagID)
	if row == nil {
		return nil, nil
	}

	var name string
	var time pgtype.Timestamp
	err := row.Scan(&name, &time)
	if err != nil {
		return nil, err
	}

	return &models.Tag{
		ID:   obj.TagID,
		Name: name,
		Time: time.Time.String(),
	}, nil
}

func (r *queryResolver) Repositories(ctx context.Context, skip *int, take *int) ([]*models.Repository, error) {
	rows, err := r.db.Query(ctx, "SELECT DISTINCT repo FROM tags OFFSET $1 LIMIT $2", skip, take)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := make([]*models.Repository, 0)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &models.Repository{
			Name: name,
		})
	}
	return repos, nil
}

func (r *queryResolver) Tags(ctx context.Context, repo string) ([]*models.Tag, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, time FROM tags WHERE repo = $1", repo)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	repos := make([]*models.Tag, 0)
	for rows.Next() {
		var id int64
		var name string
		var time pgtype.Timestamp
		err = rows.Scan(&id, &name, &time)
		if err != nil {
			return nil, err
		}
		repos = append(repos, &models.Tag{
			ID:   id,
			Name: name,
			Time: time.Time.String(),
		})
	}
	return repos, nil
}

func (r *queryResolver) Tag(ctx context.Context, repo string, name string) (*models.Tag, error) {
	row := r.db.QueryRow(ctx, "SELECT DISTINCT id, time FROM tags WHERE repo = $1 AND name = $2 LIMIT 1", repo, name)
	if row == nil {
		return nil, nil
	}

	var id int64
	var time pgtype.Timestamp
	err := row.Scan(&id, &time)
	if err != nil {
		return nil, err
	}

	return &models.Tag{
		ID:   id,
		Name: name,
		Time: time.Time.String(),
	}, nil
}

func (r *repositoryResolver) Tags(ctx context.Context, obj *models.Repository) ([]*models.Tag, error) {
	return r.Query().Tags(ctx, obj.Name)
}

func (r *tagResolver) Repo(ctx context.Context, obj *models.Tag) (*models.Repository, error) {
	row := r.db.QueryRow(ctx, "SELECT DISTINCT repo FROM tags WHERE name = $1 LIMIT 1", obj.Name)
	if row == nil {
		return nil, fmt.Errorf("no repository found for that tag: %s", obj.Name)
	}

	var name string
	err := row.Scan(&name)
	repo := &models.Repository{
		Name: name,
	}
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *tagResolver) Crds(ctx context.Context, obj *models.Tag) ([]*models.Crd, error) {
	rows, err := r.db.Query(ctx, "SELECT \"group\", version, kind, tag_id, filename, data::jsonb FROM crds WHERE tag_id = $1", obj.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	crds := make([]*models.Crd, 0)
	for rows.Next() {
		var group string
		var version string
		var kind string
		var tagId int64
		var filename string
		var data map[string]interface{}
		err = rows.Scan(&group, &version, &kind, &tagId, &filename, &data)
		if err != nil {
			return nil, err
		}
		crds = append(crds, &models.Crd{
			Group:    group,
			Version:  version,
			Kind:     kind,
			Filename: filename,
			TagID:    tagId,
			Spec:     data,
		})
	}
	return crds, nil
}

// Crd returns graphql1.CrdResolver implementation.
func (r *Resolver) Crd() graphql1.CrdResolver { return &crdResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

// Repository returns graphql1.RepositoryResolver implementation.
func (r *Resolver) Repository() graphql1.RepositoryResolver { return &repositoryResolver{r} }

// Tag returns graphql1.TagResolver implementation.
func (r *Resolver) Tag() graphql1.TagResolver { return &tagResolver{r} }

type crdResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type repositoryResolver struct{ *Resolver }
type tagResolver struct{ *Resolver }
