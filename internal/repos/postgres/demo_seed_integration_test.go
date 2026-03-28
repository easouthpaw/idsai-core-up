//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"idsai-core-up/internal/db"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDemoSeed_Integration_RichProjectFlowFixtures(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	expected := map[uuid.UUID]string{
		uuid.MustParse("11000000-0000-0000-0000-000000000007"): "REVIEW",
		uuid.MustParse("11000000-0000-0000-0000-000000000008"): "RECRUITMENT",
		uuid.MustParse("11000000-0000-0000-0000-000000000009"): "ACTIVE",
		uuid.MustParse("11000000-0000-0000-0000-000000000010"): "GRADING",
	}

	rows, err := pool.Query(ctx, `
SELECT
  p.id,
  p.status,
  COALESCE(p.professor_review_status, 'NONE') AS professor_review_status,
  (p.professor_id IS NOT NULL) AS has_professor,
  COUNT(DISTINCT CASE WHEN pm.status = 'ACTIVE' THEN pm.user_id END) AS active_members,
  COUNT(DISTINCT t.id) AS task_count,
  COUNT(DISTINCT c.id) AS criteria_count
FROM projects p
LEFT JOIN project_members pm ON pm.project_id = p.id
LEFT JOIN tasks t ON t.project_id = p.id
LEFT JOIN project_criteria c ON c.project_id = p.id
WHERE p.id IN (
  '11000000-0000-0000-0000-000000000007'::uuid,
  '11000000-0000-0000-0000-000000000008'::uuid,
  '11000000-0000-0000-0000-000000000009'::uuid,
  '11000000-0000-0000-0000-000000000010'::uuid
)
GROUP BY p.id, p.status, p.professor_review_status, p.professor_id
`)
	require.NoError(t, err)
	defer rows.Close()

	found := map[uuid.UUID]struct{}{}
	for rows.Next() {
		var (
			projectID     uuid.UUID
			status        string
			reviewStatus  string
			hasProfessor  bool
			activeMembers int
			taskCount     int
			criteriaCount int
		)
		require.NoError(t, rows.Scan(&projectID, &status, &reviewStatus, &hasProfessor, &activeMembers, &taskCount, &criteriaCount))

		expectedStatus, ok := expected[projectID]
		require.True(t, ok, "unexpected demo project returned: %s", projectID)
		require.Equal(t, expectedStatus, status)
		require.True(t, hasProfessor, "project %s should have professor", projectID)
		require.Equal(t, "ACCEPTED", reviewStatus)
		require.GreaterOrEqual(t, activeMembers, 3, "project %s should have full active team", projectID)
		require.Equal(t, 3, taskCount, "project %s should have 3 seeded tasks", projectID)
		require.Equal(t, 3, criteriaCount, "project %s should have 3 seeded criteria", projectID)
		found[projectID] = struct{}{}
	}
	require.NoError(t, rows.Err())
	require.Len(t, found, len(expected))

	statusRows, err := pool.Query(ctx, `
SELECT status, COUNT(*)
FROM tasks
WHERE project_id IN (
  '11000000-0000-0000-0000-000000000007'::uuid,
  '11000000-0000-0000-0000-000000000008'::uuid,
  '11000000-0000-0000-0000-000000000009'::uuid,
  '11000000-0000-0000-0000-000000000010'::uuid
)
GROUP BY status
`)
	require.NoError(t, err)
	defer statusRows.Close()

	taskStatuses := map[string]int{}
	for statusRows.Next() {
		var status string
		var count int
		require.NoError(t, statusRows.Scan(&status, &count))
		taskStatuses[status] = count
	}
	require.NoError(t, statusRows.Err())
	require.Greater(t, taskStatuses["OPEN"], 0)
	require.Greater(t, taskStatuses["IN_PROGRESS"], 0)
	require.Greater(t, taskStatuses["DONE"], 0)

	var reviewCount int
	err = pool.QueryRow(ctx, `
SELECT COUNT(*)
FROM project_criterion_reviews
WHERE project_id = '11000000-0000-0000-0000-000000000010'::uuid
`).Scan(&reviewCount)
	require.NoError(t, err)
	require.Equal(t, 3, reviewCount)
}
