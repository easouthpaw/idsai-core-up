package projectflow

import (
	"context"
	"errors"
	"fmt"
	"idsai-core-up/internal/strutil"
	"sort"
	"strings"

	"github.com/google/uuid"
)

const (
	SystemTaskPositionTeamLeadCode = "TEAM_LEAD"
	SystemTaskPositionTeamLeadName = "Тимлид"
)

func normalizeStackCodes(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, raw := range input {
		v := strings.ToUpper(strings.TrimSpace(raw))
		if v == "" {
			continue
		}
		v = strutil.TruncateUTF8(v, 40)
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func stacksFromCodes(codes []string) []Stack {
	out := make([]Stack, 0, len(codes))
	for _, code := range codes {
		out = append(out, Stack{Code: code})
	}
	return out
}

func normalizePositionCode(code, name string) string {
	v := strings.ToUpper(strings.TrimSpace(code))
	if v == "" {
		v = strings.ToUpper(strings.TrimSpace(name))
	}
	v = strings.ReplaceAll(v, " ", "_")
	v = strings.ReplaceAll(v, "-", "_")
	v = strutil.TruncateUTF8(v, 40)
	return v
}

func isSystemTaskPositionCode(code string) bool {
	return strings.ToUpper(strings.TrimSpace(code)) == SystemTaskPositionTeamLeadCode
}

func normalizeSearchQuery(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func clampLimit(limit, max int) int {
	if limit <= 0 {
		limit = 20
	}
	if max <= 0 {
		max = 20
	}
	if limit > max {
		return max
	}
	return limit
}

func normalizeTaskAttachments(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, raw := range items {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		v = strutil.TruncateUTF8(v, 1000)
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func (s *Service) appendTaskActivity(
	ctx context.Context,
	projectID, taskID uuid.UUID,
	actorUserID *uuid.UUID,
	eventType, fromStatus, toStatus, title, comment string,
	attachments []string,
) error {
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	if eventType == "" {
		eventType = "STATUS_CHANGED"
	}
	fromStatus = strings.ToUpper(strings.TrimSpace(fromStatus))
	toStatus = strings.ToUpper(strings.TrimSpace(toStatus))
	title = strings.TrimSpace(title)
	comment = strings.TrimSpace(comment)
	comment = strutil.TruncateUTF8(comment, 3000)
	attachments = normalizeTaskAttachments(attachments)
	return s.tasksRepo.InsertTaskActivity(ctx, projectID, taskID, actorUserID, eventType, fromStatus, toStatus, title, comment, attachments)
}

func (s *Service) ensureTaskActivityAvailable(ctx context.Context) error {
	return s.tasksRepo.EnsureTaskActivityLogAvailable(ctx)
}

func (s *Service) ensurePositionExists(ctx context.Context, projectID, positionID uuid.UUID) (capacity int, err error) {
	capacity, err = s.positionsRepo.GetProjectPositionCapacity(ctx, projectID, positionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return 0, fmt.Errorf("%w: unknown position_id", ErrInvalidInput)
		}
		return 0, err
	}
	return capacity, nil
}

func (s *Service) ensureTeamLeadTaskPosition(ctx context.Context, projectID uuid.UUID) (Position, error) {
	return s.positionsRepo.EnsureProjectPosition(ctx, projectID, SystemTaskPositionTeamLeadCode, SystemTaskPositionTeamLeadName, 1)
}

func (s *Service) ensureMemberAssignablePosition(ctx context.Context, projectID, positionID uuid.UUID) error {
	position, err := s.positionsRepo.GetProjectPosition(ctx, projectID, positionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: unknown position_id", ErrInvalidInput)
		}
		return err
	}
	if isSystemTaskPositionCode(position.Code) {
		return fmt.Errorf("%w: reserved position_id", ErrInvalidInput)
	}
	return nil
}

func (s *Service) isProjectTeamLead(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	p, err := s.projectByID(ctx, projectID)
	if err != nil {
		return false, err
	}
	if p.CreatedBy == userID {
		return true, nil
	}
	return s.projectsRepo.HasProjectRole(ctx, userID, projectID, SystemTaskPositionTeamLeadCode)
}

func (s *Service) ensurePositionCapacity(ctx context.Context, projectID, positionID uuid.UUID, excludeUserID *uuid.UUID) error {
	capacity, err := s.ensurePositionExists(ctx, projectID, positionID)
	if err != nil {
		return err
	}

	occupied, err := s.membersRepo.CountActiveMembersByPosition(ctx, projectID, positionID, excludeUserID)
	if err != nil {
		return err
	}

	if occupied >= capacity {
		return ErrPositionFull
	}
	return nil
}

func (s *Service) ensureAssigneeMatchesPosition(ctx context.Context, projectID, userID, positionID uuid.UUID) error {
	position, err := s.positionsRepo.GetProjectPosition(ctx, projectID, positionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: unknown position_id", ErrInvalidInput)
		}
		return err
	}
	if isSystemTaskPositionCode(position.Code) {
		ok, err := s.isProjectTeamLead(ctx, userID, projectID)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("%w: assignee must be project team lead", ErrInvalidInput)
		}
		return nil
	}

	status, assignedPositionID, err := s.membersRepo.GetProjectMemberStatusAndPosition(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("%w: assignee is not a project member", ErrInvalidInput)
		}
		return err
	}
	if status != "ACTIVE" {
		return fmt.Errorf("%w: assignee must be ACTIVE member", ErrInvalidInput)
	}
	if assignedPositionID == nil || *assignedPositionID != positionID {
		return fmt.Errorf("%w: assignee role does not match task position", ErrInvalidInput)
	}
	return nil
}

func (s *Service) taskByID(ctx context.Context, projectID, taskID uuid.UUID) (Task, error) {
	return s.tasksRepo.GetTaskByID(ctx, projectID, taskID)
}
