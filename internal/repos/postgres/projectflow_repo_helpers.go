package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMemberRow(scanner rowScanner) (projectflow.Member, error) {
	var (
		m         projectflow.Member
		memberID  uuid.UUID
		projectID uuid.UUID
		userID    uuid.UUID
		position  *uuid.UUID
		invitedBy *uuid.UUID
	)

	if err := scanner.Scan(
		&memberID,
		&projectID,
		&userID,
		&position,
		&m.Status,
		&m.InviteComment,
		&invitedBy,
		&m.RespondedAt,
		&m.JoinedAt,
		&m.CreatedAt,
	); err != nil {
		return projectflow.Member{}, mapProjectFlowErr(err)
	}

	m.ID = memberID.String()
	m.ProjectID = projectID.String()
	m.UserID = userID.String()
	m.PositionID = uuidPtrToStringPtr(position)
	m.InvitedBy = uuidPtrToStringPtr(invitedBy)
	return m, nil
}

func uuidPtrToStringPtr(v *uuid.UUID) *string {
	if v == nil {
		return nil
	}
	s := v.String()
	return &s
}

func scanProjectRow(scanner rowScanner) (domain.Project, error) {
	var (
		p       domain.Project
		profID  *uuid.UUID
		groupID *uuid.UUID
	)
	if err := scanner.Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.Status,
		&p.IsPublic,
		&p.CreatedBy,
		&profID,
		&p.ProfessorReviewStatus,
		&p.FacultyID,
		&p.Visibility,
		&groupID,
		&p.RetakeCount,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return domain.Project{}, mapProjectFlowErr(err)
	}
	p.ProfessorID = profID
	p.GroupID = groupID
	return p, nil
}

func scanTaskRow(scanner rowScanner) (projectflow.Task, error) {
	var (
		t          projectflow.Task
		id         uuid.UUID
		pid        uuid.UUID
		positionID uuid.UUID
		assignee   *uuid.UUID
		createdBy  uuid.UUID
	)

	if err := scanner.Scan(
		&id,
		&pid,
		&t.Title,
		&t.Description,
		&positionID,
		&assignee,
		&t.Status,
		&createdBy,
		&t.DueAt,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.PositionCode,
		&t.PositionName,
	); err != nil {
		return projectflow.Task{}, mapProjectFlowErr(err)
	}

	t.ID = id.String()
	t.ProjectID = pid.String()
	t.PositionID = positionID.String()
	t.CreatedBy = createdBy.String()
	if assignee != nil {
		s := assignee.String()
		t.AssigneeUserID = &s
	}
	return t, nil
}

func normalizeTaskAttachments(items []string) []string {
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, raw := range items {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if len(v) > 1000 {
			v = v[:1000]
		}
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

func encodeStringSliceJSON(items []string) []byte {
	b, err := json.Marshal(items)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func decodeStringSliceJSON(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return normalizeTaskAttachments(out)
}

func isUndefinedRelationErr(err error, relation string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "42P01" {
		return false
	}
	if strings.EqualFold(pgErr.TableName, relation) {
		return true
	}
	return strings.Contains(strings.ToLower(pgErr.Message), strings.ToLower(relation))
}

func isUndefinedColumnErr(err error, column string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "42703" {
		return false
	}
	if strings.EqualFold(pgErr.ColumnName, column) {
		return true
	}
	return strings.Contains(strings.ToLower(pgErr.Message), strings.ToLower(column))
}

func projectFlowProjectColumns(includeRetake bool) string {
	retakeExpr := "0 AS retake_count"
	if includeRetake {
		retakeExpr = "retake_count"
	}
	return `id, title, description, status, is_public, created_by, professor_id,
       professor_review_status,
       faculty_id, visibility, group_id,
       ` + retakeExpr + `,
       created_at, updated_at`
}

func mapProjectFlowErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return projectflow.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: already exists", projectflow.ErrInvalidInput)
	}
	return err
}
