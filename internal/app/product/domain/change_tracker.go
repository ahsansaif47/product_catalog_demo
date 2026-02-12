package domain

// ChangeTracker tracks which fields have been modified
type ChangeTracker struct {
	dirtyFields map[string]bool
}

// NewChangeTracker creates a new change tracker
func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

// MarkDirty marks a field as modified
func (ct *ChangeTracker) MarkDirty(field string) {
	if ct == nil {
		return
	}
	ct.dirtyFields[field] = true
}

// Dirty returns true if the specified field is dirty
func (ct *ChangeTracker) Dirty(field string) bool {
	if ct == nil {
		return false
	}
	return ct.dirtyFields[field]
}

// HasChanges returns true if any field is dirty
func (ct *ChangeTracker) HasChanges() bool {
	if ct == nil {
		return false
	}
	return len(ct.dirtyFields) > 0
}

// Clear clears all dirty flags
func (ct *ChangeTracker) Clear() {
	if ct == nil {
		return
	}
	ct.dirtyFields = make(map[string]bool)
}

// DirtyFields returns a list of all dirty field names
func (ct *ChangeTracker) DirtyFields() []string {
	if ct == nil {
		return nil
	}

	fields := make([]string, 0, len(ct.dirtyFields))
	for field := range ct.dirtyFields {
		fields = append(fields, field)
	}
	return fields
}

// Field constants for change tracking
const (
	FieldID           = "id"
	FieldName         = "name"
	FieldDescription  = "description"
	FieldCategory     = "category"
	FieldBasePrice    = "base_price"
	FieldDiscount     = "discount"
	FieldStatus       = "status"
	FieldArchivedAt   = "archived_at"
)
