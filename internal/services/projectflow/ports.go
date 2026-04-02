package projectflow

import (
	"context"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

type StacksRepository interface {
	ReplaceProjectStacks(ctx context.Context, projectID uuid.UUID, stackCodes []string) error
	ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error)
}

type ProjectsRepository interface {
	GetProjectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error)
	IsActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error)
	HasProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) (bool, error)
	RevokeProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) error
	UpdateProject(ctx context.Context, projectID uuid.UUID, titleSet bool, title string, descriptionSet bool, description string) error
	OpenProjectRecruitment(ctx context.Context, projectID uuid.UUID) error
	ListStudentCandidates(ctx context.Context, facultyID, projectID, requesterUserID, projectOwnerID uuid.UUID, term string, limit int) ([]StudentCandidate, error)
}

type PositionsRepository interface {
	CreateProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (Position, error)
	ListProjectPositions(ctx context.Context, projectID uuid.UUID) ([]Position, error)
	GetProjectPositionCapacity(ctx context.Context, projectID, positionID uuid.UUID) (int, error)
	SumProjectPositionCapacities(ctx context.Context, projectID uuid.UUID) (int, error)
}

type MembersRepository interface {
	IsActiveStudentInFaculty(ctx context.Context, studentID, facultyID uuid.UUID) (bool, error)
	UpsertInvitedMember(ctx context.Context, projectID, studentID, invitedBy uuid.UUID, comment string) (Member, error)
	UpsertAppliedMember(ctx context.Context, projectID, userID uuid.UUID, comment string) (Member, error)
	ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]Member, error)
	CountActiveMembersByPosition(ctx context.Context, projectID, positionID uuid.UUID, excludeUserID *uuid.UUID) (int, error)
	GetProjectMemberStatusAndPosition(ctx context.Context, projectID, userID uuid.UUID) (status string, positionID *uuid.UUID, err error)
	CountActiveMembersWithPosition(ctx context.Context, projectID uuid.UUID) (int, error)
	ApproveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID, positionID *uuid.UUID) (Member, error)
	RejectProjectMemberApplication(ctx context.Context, projectID, memberUserID uuid.UUID) (Member, error)
	RemoveProjectMember(ctx context.Context, projectID, memberUserID uuid.UUID) (Member, error)
	SetActiveMemberPosition(ctx context.Context, projectID, memberUserID, positionID uuid.UUID) (Member, error)
	GetInvitedMemberPosition(ctx context.Context, projectID, userID uuid.UUID) (*uuid.UUID, error)
	RespondMemberInvite(ctx context.Context, projectID, userID uuid.UUID, accept bool) (Member, error)
	ListIncomingInvites(ctx context.Context, userID uuid.UUID, limit int) ([]IncomingInvite, error)
	ListOutgoingApplications(ctx context.Context, userID uuid.UUID, limit int) ([]OutgoingApplication, error)
}

type ProfessorsRepository interface {
	ListProfessorCandidates(ctx context.Context, facultyID uuid.UUID, term string, limit int, requesterUserID, projectOwnerID uuid.UUID) ([]ProfessorCandidate, error)
	IsActiveProfessorInFaculty(ctx context.Context, professorID, facultyID uuid.UUID) (bool, error)
	AssignProjectProfessor(ctx context.Context, projectID, professorID uuid.UUID) error
	GetProfessorCandidateByID(ctx context.Context, professorID, facultyID uuid.UUID) (ProfessorCandidate, error)
	RespondProfessorInvite(ctx context.Context, projectID, professorID uuid.UUID, accept bool) (domain.Project, error)
	ListProfessorReviewInvites(ctx context.Context, professorID uuid.UUID, term string, limit int) ([]domain.Project, error)
}

type CriteriaRepository interface {
	GetProjectCriteriaWeightSum(ctx context.Context, projectID uuid.UUID) (int, error)
	CreateProjectCriterion(ctx context.Context, projectID, userID uuid.UUID, title, description string, weight int) (Criterion, error)
	ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]Criterion, error)
	ListProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID) ([]CriterionGrade, error)
	UpsertProjectCriterionGrades(ctx context.Context, projectID, professorID uuid.UUID, items []CriterionGradeUpsert) error
	CountProjectCriteria(ctx context.Context, projectID uuid.UUID) (int, error)
	CountProjectGradedCriteria(ctx context.Context, projectID, professorID uuid.UUID) (int, error)
}

type LifecycleRepository interface {
	ActivateProject(ctx context.Context, projectID uuid.UUID) error
	CountProjectTasksSummary(ctx context.Context, projectID uuid.UUID) (total, done int, err error)
	MoveProjectToGrading(ctx context.Context, projectID uuid.UUID) error
	MoveProjectToCompleted(ctx context.Context, projectID uuid.UUID) error
	DeleteOwnedProject(ctx context.Context, projectID, ownerID uuid.UUID) error
}

type TasksRepository interface {
	CreateTask(ctx context.Context, projectID uuid.UUID, title, description string, positionID uuid.UUID, assigneeUserID *uuid.UUID, status string, createdBy uuid.UUID, dueAt *time.Time) (uuid.UUID, error)
	GetTaskByID(ctx context.Context, projectID, taskID uuid.UUID) (Task, error)
	ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]Task, error)
	GetTaskStatusAndTitle(ctx context.Context, projectID, taskID uuid.UUID) (status, title string, err error)
	UpdateTaskStatus(ctx context.Context, projectID, taskID uuid.UUID, status string) (uuid.UUID, error)
	GetTaskAssignContext(ctx context.Context, projectID, taskID uuid.UUID) (positionID uuid.UUID, prevStatus, taskTitle string, prevAssignee *uuid.UUID, err error)
	AssignTaskToUser(ctx context.Context, projectID, taskID, assigneeUserID uuid.UUID) (uuid.UUID, error)
	ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]TaskActivity, error)
	GetTaskCompleteContext(ctx context.Context, projectID, taskID uuid.UUID) (assigneeID *uuid.UUID, currentStatus, taskTitle string, err error)
	UpsertTaskSubmission(ctx context.Context, projectID, taskID, userID uuid.UUID, comment string, attachments []string) error
	MarkTaskDone(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, error)
	ClaimTask(ctx context.Context, projectID, taskID, userID uuid.UUID) error
	InsertTaskActivity(ctx context.Context, projectID, taskID uuid.UUID, actorUserID *uuid.UUID, eventType, fromStatus, toStatus, title, comment string, attachments []string) error
}

type AccessRepository interface {
	// ListProjectRoleCodes returns all PROJECT-scope role codes for a user in a project.
	ListProjectRoleCodes(ctx context.Context, userID, projectID uuid.UUID) ([]string, error)
	// ReplaceAssignableRoles atomically removes all assignable role codes and adds the wanted ones.
	ReplaceAssignableRoles(ctx context.Context, userID, projectID uuid.UUID, assignableCodes []string, wantCodes []string) error
	// GetMemberStatusAndCreator returns the member status and the project creator ID.
	GetMemberStatusAndCreator(ctx context.Context, userID, projectID uuid.UUID) (status string, creatorID uuid.UUID, err error)
}
