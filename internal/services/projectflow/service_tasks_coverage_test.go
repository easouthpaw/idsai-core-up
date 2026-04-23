package projectflow

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateTask_Validations(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	positionID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive}

	t.Run("empty title returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: activeProject, positionCapacity: 2}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.CreateTask(context.Background(), userID, projectID, "  ", "desc", positionID, nil, nil)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("inactive project returns ErrProjectNotActive", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectGrading},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.CreateTask(context.Background(), userID, projectID, "Task", "desc", positionID, nil, nil)
		require.ErrorIs(t, err, ErrProjectNotActive)
	})

	t.Run("with assignee creates assigned activity", func(t *testing.T) {
		assigneeID := userID
		taskID := uuid.New()
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:          activeProject,
			positionCapacity: 2,
			position:         Position{ID: positionID.String(), Code: "BACKEND"},
			memberStatus:     "ACTIVE",
			memberPositionID: &positionID,
			createdTaskID:    taskID,
			task:             Task{ID: taskID.String(), Title: "Build API"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		task, err := svc.CreateTask(context.Background(), userID, projectID, "Build API", "desc", positionID, &assigneeID, nil)
		require.NoError(t, err)
		require.Equal(t, taskID.String(), task.ID)
		// both CREATED and ASSIGNED activities should be appended
		require.Equal(t, "ASSIGNED", repo.insertedEventType)
	})
}

func TestUpdateTaskStatus_EdgeCases(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	taskID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive}

	t.Run("invalid status returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{project: activeProject}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.UpdateTaskStatus(context.Background(), userID, projectID, taskID, "UNKNOWN")
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("same status does not log activity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:       activeProject,
			taskStatus:    "OPEN",
			taskTitle:     "Build API",
			updatedTaskID: taskID,
			task:          Task{ID: taskID.String(), Status: "OPEN"},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.UpdateTaskStatus(context.Background(), userID, projectID, taskID, "open")
		require.NoError(t, err)
		require.Empty(t, repo.insertedEventType) // no activity for same-status update
	})
}

func TestAssignTask_SameAssigneeSkipsActivity(t *testing.T) {
	projectID := uuid.New()
	callerID := uuid.New()
	assigneeID := uuid.New()
	taskID := uuid.New()
	positionID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive, CreatedBy: callerID}

	t.Run("same assignee does not log ASSIGNED activity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:           activeProject,
			assignPositionID:  positionID,
			assignPrevStatus:  "OPEN",
			assignTaskTitle:   "Build API",
			assignPrevAssignee: &assigneeID, // same as new assignee
			assignedTaskID:    taskID,
			position:          Position{ID: positionID.String(), Code: "BACKEND"},
			memberStatus:      "ACTIVE",
			memberPositionID:  &positionID,
			task:              Task{ID: taskID.String()},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.AssignTask(context.Background(), callerID, projectID, taskID, assigneeID)
		require.NoError(t, err)
		require.Empty(t, repo.insertedEventType)
	})
}

func TestCompleteTask_Guards(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	taskID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive}

	t.Run("non-assignee returns ErrForbidden", func(t *testing.T) {
		otherID := uuid.New()
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:            activeProject,
			completeAssigneeID: &otherID, // different user owns the task
			completeStatus:     "IN_PROGRESS",
			completeTitle:      "Build API",
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.CompleteTask(context.Background(), userID, projectID, taskID, "done", nil)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("task not in progress returns ErrInvalidInput", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:            activeProject,
			completeAssigneeID: &userID,
			completeStatus:     "OPEN",
			completeTitle:      "Build API",
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.CompleteTask(context.Background(), userID, projectID, taskID, "done", nil)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("no assignee returns ErrForbidden", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:            activeProject,
			completeAssigneeID: nil, // unassigned
			completeStatus:     "IN_PROGRESS",
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		_, err := svc.CompleteTask(context.Background(), userID, projectID, taskID, "done", nil)
		require.ErrorIs(t, err, domain.ErrForbidden)
	})
}

func TestClaimTask_RecordsActivity(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	taskID := uuid.New()
	activeProject := domain.Project{ID: projectID, Status: domain.ProjectActive}

	t.Run("success records CLAIMED activity", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project:    activeProject,
			taskStatus: "OPEN",
			taskTitle:  "Build API",
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		err := svc.ClaimTask(context.Background(), userID, projectID, taskID)
		require.NoError(t, err)
		require.Equal(t, "CLAIMED", repo.insertedEventType)
		require.Equal(t, "IN_PROGRESS", repo.insertedToStatus)
	})
}

func TestDeleteTask_Success(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	taskID := uuid.New()

	t.Run("success deletes task", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectActive},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		require.NoError(t, svc.DeleteTask(context.Background(), userID, projectID, taskID))
	})

	t.Run("inactive project returns ErrProjectNotActive", func(t *testing.T) {
		authz := &projectFlowAuthz{canResult: true}
		repo := &projectFlowRepos{
			project: domain.Project{ID: projectID, Status: domain.ProjectCompleted},
		}
		svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
		require.ErrorIs(t, svc.DeleteTask(context.Background(), userID, projectID, taskID), ErrProjectNotActive)
	})
}

func TestListTasks_NoPermission(t *testing.T) {
	authz := &projectFlowAuthz{canResult: false}
	repo := &projectFlowRepos{}
	svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
	_, err := svc.ListTasks(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestListTaskActivities_NoPermission(t *testing.T) {
	authz := &projectFlowAuthz{canResult: false}
	repo := &projectFlowRepos{}
	svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
	_, err := svc.ListTaskActivities(context.Background(), uuid.New(), uuid.New(), nil)
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateTask_DueAtPropagated(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	positionID := uuid.New()
	taskID := uuid.New()
	dueAt := time.Now().Add(24 * time.Hour)

	authz := &projectFlowAuthz{canResult: true}
	repo := &projectFlowRepos{
		project:          domain.Project{ID: projectID, Status: domain.ProjectActive},
		positionCapacity: 2,
		createdTaskID:    taskID,
		task:             Task{ID: taskID.String(), Title: "Deadline Task"},
	}
	svc := newProjectFlowService(repo, authz, &projectFlowGrantor{})
	task, err := svc.CreateTask(context.Background(), userID, projectID, "Deadline Task", "", positionID, nil, &dueAt)
	require.NoError(t, err)
	require.Equal(t, taskID.String(), task.ID)
}
