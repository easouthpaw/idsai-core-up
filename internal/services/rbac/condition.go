package rbac

import "sync"

// ConditionFunc validates runtime attributes for a permission check.
// It receives a map of contextual attributes (e.g. user_id, task_author_id,
// project_status) and returns true if the condition is satisfied.
type ConditionFunc func(attrs map[string]interface{}) bool

// ConditionRegistry holds ABAC condition functions keyed by permission code.
// Multiple conditions can be registered per permission; ALL must pass.
type ConditionRegistry struct {
	mu    sync.RWMutex
	funcs map[string][]ConditionFunc
}

// NewConditionRegistry creates an empty registry.
func NewConditionRegistry() *ConditionRegistry {
	return &ConditionRegistry{
		funcs: make(map[string][]ConditionFunc),
	}
}

// Register adds a condition function for the given permission code.
// If multiple functions are registered for the same code, ALL must pass.
func (r *ConditionRegistry) Register(permissionCode string, fn ConditionFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funcs[permissionCode] = append(r.funcs[permissionCode], fn)
}

// Evaluate checks all registered conditions for a permission code.
// Returns true if no conditions are registered or ALL pass.
func (r *ConditionRegistry) Evaluate(permissionCode string, attrs map[string]interface{}) bool {
	r.mu.RLock()
	fns, ok := r.funcs[permissionCode]
	r.mu.RUnlock()

	if !ok || len(fns) == 0 {
		return true // no conditions — always pass
	}

	for _, fn := range fns {
		if !fn(attrs) {
			return false
		}
	}
	return true
}

// Has returns true if at least one condition is registered for the code.
func (r *ConditionRegistry) Has(permissionCode string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fns, ok := r.funcs[permissionCode]
	return ok && len(fns) > 0
}
