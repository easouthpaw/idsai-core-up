package projectflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/strutil"

	"github.com/google/uuid"
)

func (s *Service) ensureActiveProject(ctx context.Context, projectID uuid.UUID) error {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return err
	}
	if p.Status != domain.ProjectActive {
		return ErrProjectNotActive
	}
	return nil
}

func normalizeTaskStatus(status string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(status))
	if s == "OPEN" || s == "IN_PROGRESS" || s == "DONE" {
		return s, true
	}
	return "", false
}

func (s *Service) CreateTask(ctx context.Context, userID, projectID uuid.UUID, title, description string, positionID uuid.UUID, assigneeUserID *uuid.UUID, dueAt *time.Time) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.create", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return Task{}, ErrInvalidInput
	}
	if _, err := s.ensurePositionExists(ctx, projectID, positionID); err != nil {
		return Task{}, err
	}
	if assigneeUserID != nil {
		if err := s.ensureAssigneeMatchesPosition(ctx, projectID, *assigneeUserID, positionID); err != nil {
			return Task{}, err
		}
	}
	if err := s.ensureTaskActivityAvailable(ctx); err != nil {
		return Task{}, err
	}

	status := "OPEN"
	if repo, ok := s.tasksRepo.(AtomicTasksRepository); ok {
		taskID, err := repo.CreateTaskWithActivity(ctx, projectID, title, description, positionID, assigneeUserID, status, userID, dueAt)
		if err != nil {
			return Task{}, err
		}
		return s.taskByID(ctx, projectID, taskID)
	}

	taskID, err := s.tasksRepo.CreateTask(ctx, projectID, title, description, positionID, assigneeUserID, status, userID, dueAt)
	if err != nil {
		return Task{}, err
	}
	if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "CREATED", "", status, title, description, nil); err != nil {
		return Task{}, err
	}
	if assigneeUserID != nil {
		if err := s.appendTaskActivity(
			ctx,
			projectID,
			taskID,
			&userID,
			"ASSIGNED",
			status,
			status,
			title,
			fmt.Sprintf("Назначен исполнитель: %s", assigneeUserID.String()),
			nil,
		); err != nil {
			return Task{}, err
		}
	}
	return s.taskByID(ctx, projectID, taskID)
}

func (s *Service) ListTasks(ctx context.Context, userID, projectID uuid.UUID) ([]Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.view", projectID); err != nil {
		return nil, err
	}
	return s.tasksRepo.ListProjectTasks(ctx, projectID)
}

func (s *Service) UpdateTaskStatus(ctx context.Context, userID, projectID, taskID uuid.UUID, status string) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.update", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	status, ok := normalizeTaskStatus(status)
	if !ok {
		return Task{}, ErrInvalidInput
	}
	if err := s.ensureTaskActivityAvailable(ctx); err != nil {
		return Task{}, err
	}

	prevStatus, taskTitle, err := s.tasksRepo.GetTaskStatusAndTitle(ctx, projectID, taskID)
	if err != nil {
		return Task{}, err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))
	outTaskID, err := s.tasksRepo.UpdateTaskStatus(ctx, projectID, taskID, status)
	if err != nil {
		return Task{}, err
	}
	if prevStatus != status {
		if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "STATUS_CHANGED", prevStatus, status, taskTitle, "", nil); err != nil {
			return Task{}, err
		}
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) AssignTask(ctx context.Context, userID, projectID, taskID, assigneeUserID uuid.UUID) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.assign", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureTaskActivityAvailable(ctx); err != nil {
		return Task{}, err
	}

	positionID, prevStatus, taskTitle, prevAssignee, err := s.tasksRepo.GetTaskAssignContext(ctx, projectID, taskID)
	if err != nil {
		return Task{}, err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))
	if err := s.ensureAssigneeMatchesPosition(ctx, projectID, assigneeUserID, positionID); err != nil {
		return Task{}, err
	}
	outTaskID, err := s.tasksRepo.AssignTaskToUser(ctx, projectID, taskID, assigneeUserID)
	if err != nil {
		return Task{}, err
	}
	if prevAssignee == nil || *prevAssignee != assigneeUserID {
		if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "ASSIGNED", prevStatus, prevStatus, taskTitle, fmt.Sprintf("Назначен исполнитель: %s", assigneeUserID.String()), nil); err != nil {
			return Task{}, err
		}
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) ListTaskActivities(ctx context.Context, userID, projectID uuid.UUID, taskID *uuid.UUID) ([]TaskActivity, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.view", projectID); err != nil {
		return nil, err
	}
	return s.tasksRepo.ListProjectTaskActivities(ctx, projectID, taskID)
}

func (s *Service) CompleteTask(ctx context.Context, userID, projectID, taskID uuid.UUID, comment string, attachments []string) (Task, error) {
	if err := s.requireProjectPermission(ctx, userID, "task.update", projectID); err != nil {
		return Task{}, err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return Task{}, err
	}
	comment = strings.TrimSpace(comment)
	comment = strutil.TruncateUTF8(comment, 3000)
	attachments = normalizeTaskAttachments(attachments)

	assigneeID, currentStatus, taskTitle, err := s.tasksRepo.GetTaskCompleteContext(ctx, projectID, taskID)
	if err != nil {
		return Task{}, err
	}
	if assigneeID == nil || *assigneeID != userID {
		return Task{}, domain.ErrForbidden
	}
	currentStatus = strings.ToUpper(strings.TrimSpace(currentStatus))
	if currentStatus != "IN_PROGRESS" {
		return Task{}, ErrInvalidInput
	}
	if err := s.ensureTaskActivityAvailable(ctx); err != nil {
		return Task{}, err
	}

	var outTaskID uuid.UUID
	if repo, ok := s.tasksRepo.(AtomicTasksRepository); ok {
		outTaskID, err = repo.CompleteTaskWithSubmission(ctx, projectID, taskID, userID, taskTitle, comment, attachments)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return Task{}, fmt.Errorf("%w: task was already completed by a concurrent request", ErrInvalidInput)
			}
			return Task{}, err
		}
	} else {
		if err := s.tasksRepo.UpsertTaskSubmission(ctx, projectID, taskID, userID, comment, attachments); err != nil {
			return Task{}, err
		}
		outTaskID, err = s.tasksRepo.MarkTaskDone(ctx, projectID, taskID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return Task{}, fmt.Errorf("%w: task was already completed by a concurrent request", ErrInvalidInput)
			}
			return Task{}, err
		}

		if err := s.appendTaskActivity(ctx, projectID, taskID, &userID, "COMPLETED", "IN_PROGRESS", "DONE", taskTitle, comment, attachments); err != nil {
			return Task{}, err
		}
	}
	return s.taskByID(ctx, projectID, outTaskID)
}

func (s *Service) ClaimTask(ctx context.Context, userID, projectID, taskID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "task.claim", projectID); err != nil {
		return err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return err
	}

	prevStatus, taskTitle, err := s.tasksRepo.GetTaskStatusAndTitle(ctx, projectID, taskID)
	if err != nil {
		return err
	}
	prevStatus = strings.ToUpper(strings.TrimSpace(prevStatus))
	if err := s.ensureTaskActivityAvailable(ctx); err != nil {
		return err
	}
	if repo, ok := s.tasksRepo.(AtomicTasksRepository); ok {
		return repo.ClaimTaskWithActivity(ctx, projectID, taskID, userID, prevStatus, taskTitle)
	}
	if err := s.tasksRepo.ClaimTask(ctx, projectID, taskID, userID); err != nil {
		return err
	}
	return s.appendTaskActivity(ctx, projectID, taskID, &userID, "CLAIMED", prevStatus, "IN_PROGRESS", taskTitle, "", nil)
}

func (s *Service) DeleteTask(ctx context.Context, userID, projectID, taskID uuid.UUID) error {
	if err := s.requireProjectPermission(ctx, userID, "task.delete", projectID); err != nil {
		return err
	}
	if err := s.ensureActiveProject(ctx, projectID); err != nil {
		return err
	}
	return s.tasksRepo.DeleteTask(ctx, projectID, taskID)
}
