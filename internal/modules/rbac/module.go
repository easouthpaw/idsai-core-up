package rbacmodule

import (
	"time"

	"idsai-core-up/internal/infra/cache"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/rbac"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Output struct {
	Repo       *postgres.RBACRepo
	Service    *rbac.Service
	Authorizer rbac.Authorizer
}

func New(pool *pgxpool.Pool, redisClient *cache.RedisClient) Output {
	repo := postgres.NewRBACRepo(pool)
	conditions := rbac.NewConditionRegistry()

	// Example ABAC condition: task.edit allowed only if
	// user_id == task_author_id AND project_status != "COMPLETED"
	conditions.Register("task.edit", func(attrs map[string]interface{}) bool {
		userID, _ := attrs["user_id"].(string)
		authorID, _ := attrs["task_author_id"].(string)
		status, _ := attrs["project_status"].(string)
		return userID == authorID && status != "COMPLETED"
	})

	svc := rbac.NewService(repo, conditions)

	var authorizer rbac.Authorizer = svc
	if redisClient != nil {
		cached := rbac.NewCachedAuthorizer(svc, redisClient, 10*time.Minute)
		authorizer = cached
		// Wire cache invalidation into the repo layer
		repo.SetCacheInvalidator(cached.InvalidateUser)
	}

	return Output{
		Repo:       repo,
		Service:    svc,
		Authorizer: authorizer,
	}
}
