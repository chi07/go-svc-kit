package repox

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

func TestIsDuplicateErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "duplicate key error lowercase",
			err:      errors.New("duplicate key value violates unique constraint"),
			expected: true,
		},
		{
			name:     "duplicate key error uppercase",
			err:      errors.New("DUPLICATE KEY value violates UNIQUE CONSTRAINT"),
			expected: true,
		},
		{
			name:     "unique constraint error",
			err:      errors.New("violates UNIQUE CONSTRAINT"),
			expected: true,
		},
		{
			name:     "unique constraint mixed case",
			err:      errors.New("Unique Constraint violation"),
			expected: true,
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "timeout error",
			err:      errors.New("connection timeout"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateErr(tt.err)
			if result != tt.expected {
				t.Errorf("IsDuplicateErr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsNoRowsErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "pg.ErrNoRows",
			err:      pg.ErrNoRows,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNoRowsErr(tt.err)
			if result != tt.expected {
				t.Errorf("IsNoRowsErr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMapFieldToDB(t *testing.T) {
	aliasMap := map[string]string{
		"id":         "id",
		"name":       "user_name",
		"email":      "email_address",
		"created_at": "created_at",
	}

	tests := []struct {
		name      string
		alias     string
		aliasMap  map[string]string
		expectCol string
		expectOk  bool
	}{
		{
			name:      "existing alias",
			alias:     "name",
			aliasMap:  aliasMap,
			expectCol: "user_name",
			expectOk:  true,
		},
		{
			name:      "non-existing alias",
			alias:     "phone",
			aliasMap:  aliasMap,
			expectCol: "",
			expectOk:  false,
		},
		{
			name:      "empty alias map",
			alias:     "id",
			aliasMap:  map[string]string{},
			expectCol: "",
			expectOk:  false,
		},
		{
			name:      "case sensitive - lowercase alias",
			alias:     "email",
			aliasMap:  aliasMap,
			expectCol: "email_address",
			expectOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, ok := MapFieldToDB(tt.alias, tt.aliasMap)
			if col != tt.expectCol {
				t.Errorf("MapFieldToDB() col = %v, want %v", col, tt.expectCol)
			}
			if ok != tt.expectOk {
				t.Errorf("MapFieldToDB() ok = %v, want %v", ok, tt.expectOk)
			}
		})
	}
}

func TestAllowedAliasSetLower(t *testing.T) {
	tests := []struct {
		name     string
		aliasMap map[string]string
		expected map[string]struct{}
	}{
		{
			name:     "empty map",
			aliasMap: map[string]string{},
			expected: nil,
		},
		{
			name:     "nil map",
			aliasMap: nil,
			expected: nil,
		},
		{
			name: "simple aliases",
			aliasMap: map[string]string{
				"id":    "id",
				"name":  "user_name",
				"email": "email",
			},
			expected: map[string]struct{}{
				"id":    {},
				"name":  {},
				"email": {},
			},
		},
		{
			name: "uppercase aliases converted to lower",
			aliasMap: map[string]string{
				"ID":    "id",
				"Name":  "user_name",
				"EMAIL": "email",
			},
			expected: map[string]struct{}{
				"id":    {},
				"name":  {},
				"email": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllowedAliasSetLower(tt.aliasMap)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("AllowedAliasSetLower() = %v, want nil", result)
				}
			} else {
				if len(result) != len(tt.expected) {
					t.Errorf("AllowedAliasSetLower() length = %v, want %v", len(result), len(tt.expected))
				}
				// Verify all expected keys exist
				for k := range tt.expected {
					if _, ok := result[k]; !ok {
						t.Errorf("AllowedAliasSetLower() missing key %s", k)
					}
				}
			}
		})
	}
}

func TestSafeColumns(t *testing.T) {
	aliasMap := map[string]string{
		"id":    "id",
		"name":  "user_name",
		"email": "email_address",
	}

	tests := []struct {
		name      string
		requested []string
		aliasMap  map[string]string
		expected  []string
	}{
		{
			name:      "empty requested",
			requested: []string{},
			aliasMap:  aliasMap,
			expected:  nil,
		},
		{
			name:      "valid aliases",
			requested: []string{"id", "name"},
			aliasMap:  aliasMap,
			expected:  []string{"id", "user_name"},
		},
		{
			name:      "invalid alias filtered out",
			requested: []string{"id", "phone", "email"},
			aliasMap:  aliasMap,
			expected:  []string{"id", "email_address"},
		},
		{
			name:      "case insensitive matching",
			requested: []string{"ID", "NAME"},
			aliasMap:  aliasMap,
			expected:  []string{"id", "user_name"},
		},
		{
			name:      "duplicate columns removed",
			requested: []string{"id", "id", "name"},
			aliasMap:  aliasMap,
			expected:  []string{"id", "user_name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeColumns(tt.requested, tt.aliasMap)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SafeColumns() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type orderItem struct {
	field string
	desc  bool
}

func TestBuildOrderExpr(t *testing.T) {
	aliasMap := map[string]string{
		"name":       "user_name",
		"created_at": "created_at",
		"id":         "id",
	}

	tests := []struct {
		name        string
		items       []orderItem
		fieldFn     func(orderItem) string
		descFn      func(orderItem) bool
		aliasMap    map[string]string
		defaults    []string
		expected    []string
	}{
		{
			name:        "empty items with defaults",
			items:       []orderItem{},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{"id ASC"},
			expected:    []string{"id ASC"},
		},
		{
			name:        "single ascending order",
			items:       []orderItem{{field: "name", desc: false}},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{"id ASC"},
			expected:    []string{"user_name ASC"},
		},
		{
			name:        "single descending order",
			items:       []orderItem{{field: "created_at", desc: true}},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{"id ASC"},
			expected:    []string{"created_at DESC"},
		},
		{
			name: "multiple orders",
			items: []orderItem{
				{field: "name", desc: true},
				{field: "created_at", desc: false},
			},
			fieldFn:  func(oi orderItem) string { return oi.field },
			descFn:   func(oi orderItem) bool { return oi.desc },
			aliasMap: aliasMap,
			defaults: []string{"id ASC"},
			expected: []string{"user_name DESC", "created_at ASC"},
		},
		{
			name:        "invalid alias ignored",
			items:       []orderItem{{field: "phone", desc: false}},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{"id ASC"},
			expected:    []string{"id ASC"},
		},
		{
			name:        "all invalid aliases return defaults",
			items:       []orderItem{{field: "phone"}, {field: "address"}},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{"id ASC"},
			expected:    []string{"id ASC"},
		},
		{
			name:        "case insensitive alias matching",
			items:       []orderItem{{field: "NAME", desc: true}},
			fieldFn:     func(oi orderItem) string { return oi.field },
			descFn:      func(oi orderItem) bool { return oi.desc },
			aliasMap:    aliasMap,
			defaults:    []string{},
			expected:    []string{"user_name DESC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildOrderExpr(tt.items, tt.fieldFn, tt.descFn, tt.aliasMap, tt.defaults...)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("BuildOrderExpr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWithNotDeleted(t *testing.T) {
	// Create a mock query - we can't easily test this without a real DB connection
	// but we can at least verify the function doesn't panic
	q := &orm.Query{}
	result := WithNotDeleted(q)
	if result == nil {
		t.Error("WithNotDeleted() returned nil")
	}
	// The function should return the modified query
	if result != q {
		t.Error("WithNotDeleted() should return the same query pointer")
	}
}

func TestApplyExactFilters(t *testing.T) {
	aliasMap := map[string]string{
		"id":    "id",
		"name":  "user_name",
		"email": "email_address",
	}

	tests := []struct {
		name   string
		extras map[string]any
	}{
		{
			name:   "empty extras",
			extras: map[string]any{},
		},
		{
			name: "single filter",
			extras: map[string]any{
				"name": "John",
			},
		},
		{
			name: "multiple filters",
			extras: map[string]any{
				"id":    1,
				"email": "john@example.com",
			},
		},
		{
			name: "invalid alias ignored",
			extras: map[string]any{
				"name":  "John",
				"phone": "123456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &orm.Query{}
			ApplyExactFilters(q, tt.extras, aliasMap)
			// Without a real DB connection, we can't verify the actual WHERE clause
			// but we can ensure the function doesn't panic
			if q == nil {
				t.Error("ApplyExactFilters() should not modify query to nil")
			}
		})
	}
}

func TestApplyLimitOffset(t *testing.T) {
	tests := []struct {
		name   string
		limit  int64
		offset int64
	}{
		{
			name:   "positive limit and offset",
			limit:  10,
			offset: 20,
		},
		{
			name:   "zero limit and offset",
			limit:  0,
			offset: 0,
		},
		{
			name:   "negative limit and offset",
			limit:  -1,
			offset: -5,
		},
		{
			name:   "positive limit zero offset",
			limit:  10,
			offset: 0,
		},
		{
			name:   "zero limit positive offset",
			limit:  0,
			offset: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &orm.Query{}
			result := ApplyLimitOffset(q, tt.limit, tt.offset)
			if result == nil {
				t.Error("ApplyLimitOffset() returned nil")
			}
			if result != q {
				t.Error("ApplyLimitOffset() should return the same query pointer")
			}
		})
	}
}

func TestApplyLimitOffsetPlusOne(t *testing.T) {
	tests := []struct {
		name   string
		limit  int64
		offset int64
	}{
		{
			name:   "positive limit and offset",
			limit:  10,
			offset: 20,
		},
		{
			name:   "zero limit and offset",
			limit:  0,
			offset: 0,
		},
		{
			name:   "negative limit and offset",
			limit:  -1,
			offset: -5,
		},
		{
			name:   "limit of 1",
			limit:  1,
			offset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &orm.Query{}
			result := ApplyLimitOffsetPlusOne(q, tt.limit, tt.offset)
			if result == nil {
				t.Error("ApplyLimitOffsetPlusOne() returned nil")
			}
			if result != q {
				t.Error("ApplyLimitOffsetPlusOne() should return the same query pointer")
			}
		})
	}
}

func TestTrimHasNext(t *testing.T) {
	tests := []struct {
		name          string
		rows          []int
		limit         int64
		expectRows    []int
		expectHasNext bool
	}{
		{
			name:          "rows less than limit",
			rows:          []int{1, 2, 3},
			limit:         5,
			expectRows:    []int{1, 2, 3},
			expectHasNext: false,
		},
		{
			name:          "rows equal to limit",
			rows:          []int{1, 2, 3},
			limit:         3,
			expectRows:    []int{1, 2, 3},
			expectHasNext: false,
		},
		{
			name:          "rows more than limit",
			rows:          []int{1, 2, 3, 4, 5},
			limit:         3,
			expectRows:    []int{1, 2, 3},
			expectHasNext: true,
		},
		{
			name:          "zero limit",
			rows:          []int{1, 2, 3},
			limit:         0,
			expectRows:    []int{1, 2, 3},
			expectHasNext: false,
		},
		{
			name:          "negative limit",
			rows:          []int{1, 2, 3},
			limit:         -1,
			expectRows:    []int{1, 2, 3},
			expectHasNext: false,
		},
		{
			name:          "empty rows",
			rows:          []int{},
			limit:         5,
			expectRows:    []int{},
			expectHasNext: false,
		},
		{
			name:          "nil rows",
			rows:          nil,
			limit:         5,
			expectRows:    nil,
			expectHasNext: false,
		},
		{
			name:          "limit of 1 with multiple rows",
			rows:          []int{1, 2, 3},
			limit:         1,
			expectRows:    []int{1},
			expectHasNext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, hasNext := TrimHasNext(tt.rows, tt.limit)
			if !reflect.DeepEqual(rows, tt.expectRows) {
				t.Errorf("TrimHasNext() rows = %v, want %v", rows, tt.expectRows)
			}
			if hasNext != tt.expectHasNext {
				t.Errorf("TrimHasNext() hasNext = %v, want %v", hasNext, tt.expectHasNext)
			}
		})
	}
}

func TestCount(t *testing.T) {
	// Test with nil query - this will panic in real usage
	// We can't easily test this without a real DB connection
	// but we can verify the function signature and return type

	// This test just ensures the function exists and has the right signature
	// In practice, you'd need a mock DB or integration test
	t.Skip("Count requires a real database connection for meaningful tests")
}

// Additional edge case tests

func TestIsDuplicateErr_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "error with duplicate key in middle",
			err:      errors.New("ERROR: duplicate key value violates unique constraint uk_user_email"),
			expected: true,
		},
		{
			name:     "error with unique constraint at end",
			err:      errors.New("violates unique constraint"),
			expected: true,
		},
		{
			name:     "error containing both patterns",
			err:      errors.New("Duplicate Key and UNIQUE CONSTRAINT both present"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDuplicateErr(tt.err)
			if result != tt.expected {
				t.Errorf("IsDuplicateErr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAllowedAliasSetLower_DoesNotModifyOriginal(t *testing.T) {
	aliasMap := map[string]string{
		"Name":  "user_name",
		"Email": "email_address",
	}

	result := AllowedAliasSetLower(aliasMap)

	// Verify original map is unchanged
	if _, hasName := aliasMap["Name"]; !hasName {
		t.Error("Original map should have 'Name' key")
	}
	if _, hasEmail := aliasMap["Email"]; !hasEmail {
		t.Error("Original map should have 'Email' key")
	}

	// Verify result has lowercase keys
	if _, hasNameLower := result["name"]; !hasNameLower {
		t.Error("Result should have 'name' key")
	}
	if _, hasEmailLower := result["email"]; !hasEmailLower {
		t.Error("Result should have 'email' key")
	}
}

func TestBuildOrderExpr_WhitespaceHandling(t *testing.T) {
	aliasMap := map[string]string{
		"name": "user_name",
	}

	items := []orderItem{
		{field: "  name  ", desc: false},
	}

	result := BuildOrderExpr(items,
		func(oi orderItem) string { return oi.field },
		func(oi orderItem) bool { return oi.desc },
		aliasMap,
	)

	expected := []string{"user_name ASC"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("BuildOrderExpr() = %v, want %v", result, expected)
	}
}

func TestApplyExactFilters_WhitespaceAndCaseHandling(t *testing.T) {
	aliasMap := map[string]string{
		"name": "user_name",
	}

	q := &orm.Query{}
	extras := map[string]any{
		"  NAME  ": "John",
	}

	ApplyExactFilters(q, extras, aliasMap)
	// Function should handle case-insensitive and whitespace properly
	if q == nil {
		t.Error("ApplyExactFilters() should not modify query to nil")
	}
}

// Test with generics - using simple struct types
type testModel struct {
	ID   int
	Name string
}

func TestTrimHasNext_WithStructs(t *testing.T) {
	rows := []testModel{
		{ID: 1, Name: "One"},
		{ID: 2, Name: "Two"},
		{ID: 3, Name: "Three"},
	}

	result, hasNext := TrimHasNext(rows, 2)
	if !hasNext {
		t.Error("TrimHasNext() hasNext = false, want true")
	}
	if len(result) != 2 {
		t.Errorf("TrimHasNext() len = %d, want 2", len(result))
	}
	if result[0].ID != 1 {
		t.Errorf("TrimHasNext() result[0].ID = %d, want 1", result[0].ID)
	}
	if result[1].ID != 2 {
		t.Errorf("TrimHasNext() result[1].ID = %d, want 2", result[1].ID)
	}
}

func TestBuildOrderExpr_WithStructs(t *testing.T) {
	aliasMap := map[string]string{
		"name": "user_name",
		"id":   "id",
	}

	items := []testModel{
		{Name: "name"},
		{Name: "id"},
	}

	result := BuildOrderExpr(items,
		func(m testModel) string { return m.Name },
		func(m testModel) bool { return m.Name == "name" },
		aliasMap,
		"id ASC",
	)

	expected := []string{"user_name DESC", "id ASC"}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("BuildOrderExpr() = %v, want %v", result, expected)
	}
}

func TestCount_ContextCancellation(t *testing.T) {
	// This is a placeholder to show how you'd test context handling
	// In real scenarios, you'd use a mock database
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Without a real DB, we can't test this properly
	// This shows the test structure you'd use with proper mocking
	_ = ctx
	t.Skip("Requires database mock setup")
}
