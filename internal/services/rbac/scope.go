package rbac

import "github.com/google/uuid"

type ScopeType string

const (
	ScopeSystem  ScopeType = "SYSTEM"
	ScopeFaculty ScopeType = "FACULTY"
	ScopeProject ScopeType = "PROJECT"
)

type Scope struct {
	Type ScopeType
	ID   *uuid.UUID
}

func (s Scope) Validate() bool {
	if s.Type == ScopeSystem {
		return s.ID == nil
	}
	if s.Type == ScopeFaculty || s.Type == ScopeProject {
		return s.ID != nil
	}
	return false
}
