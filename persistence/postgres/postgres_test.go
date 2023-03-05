package postgres

import (
	"fmt"
	"reflect"
	"testing"
)

func TestBuildRestoreQuery(t *testing.T) {
	tests := []struct {
		kind           string
		identity       string
		expectedQuery  string
		expectedParams []any
	}{
		{
			kind:           "",
			identity:       "",
			expectedQuery:  "SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots ORDER BY kind, identity, created_at DESC",
			expectedParams: make([]any, 0),
		},
		{
			kind:           "inventory",
			identity:       "",
			expectedQuery:  "SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots WHERE kind = $1 ORDER BY kind, identity, created_at DESC",
			expectedParams: []any{"inventory"},
		},
		{
			kind:           "",
			identity:       "test",
			expectedQuery:  "SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots WHERE identity = $1 ORDER BY kind, identity, created_at DESC",
			expectedParams: []any{"test"},
		},
		{
			kind:           "inventory",
			identity:       "test",
			expectedQuery:  "SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots WHERE kind = $1 AND identity = $2 ORDER BY kind, identity, created_at DESC",
			expectedParams: []any{"inventory", "test"},
		},
	}

	for i, tt := range tests {
		tf := func(t *testing.T) {
			actualQuery, actualParams := buildRestoreQuery(tt.kind, tt.identity)

			if actualQuery != tt.expectedQuery {
				t.Fatalf("FAIL: expected %s, got %s", tt.expectedQuery, actualQuery)
			}

			if !reflect.DeepEqual(tt.expectedParams, actualParams) {
				t.Fatalf("FAIL: expected %v, got %v", tt.expectedParams, actualParams)
			}
		}
		t.Run(fmt.Sprintf("case_%d", i), tf)
	}
}
