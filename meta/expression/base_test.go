package expression

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockReducible struct {
	val string
}

func (m mockReducible) Reduce(resolver NameResolver) any {
	return resolver(m.val)
}

type mockBase struct {
	val any
}

func (m mockBase) ToRepr() any {
	return m.val
}

// A mock that is both Reducible and Base, but we expect Reduce to be called first
type mockBoth struct {
	val string
}

func (m mockBoth) ToRepr() any {
	return "BASE_" + m.val
}

func (m mockBoth) Reduce(resolver NameResolver) any {
	return "REDUCED_" + resolver(m.val)
}

func TestNormalizeExpression(t *testing.T) {
	resolver := func(s string) string {
		return "resolved_" + s
	}

	tests := []struct {
		name     string
		expr     any
		expected any
	}{
		{
			name:     "simple type",
			expr:     123,
			expected: 123,
		},
		{
			name:     "Reducible",
			expr:     mockReducible{val: "field"},
			expected: "resolved_field",
		},
		{
			name:     "Base",
			expr:     mockBase{val: "base_val"},
			expected: "base_val",
		},
		{
			name:     "Both Reducible and Base",
			expr:     mockBoth{val: "field"},
			expected: "REDUCED_resolved_field", // Reduce takes precedence
		},
		{
			name:     "FieldName",
			expr:     FieldName{Name: "field"},
			expected: "resolved_field",
		},
		{
			name:     "LiteralValue",
			expr:     LiteralValue{Value: 42, Context: FieldName{Name: "field"}},
			expected: 42,
		},
		{
			name: "bson.A simple",
			expr: bson.A{1, "test"},
			expected: bson.A{1, "test"},
		},
		{
			name: "bson.A with nested Reducible",
			expr: bson.A{1, mockReducible{val: "field"}},
			expected: bson.A{1, "resolved_field"},
		},
		{
			name: "bson.D simple",
			expr: bson.D{kv("key", "val")},
			expected: bson.D{kv("key", "val")},
		},
		{
			name: "bson.D with nested Reducible",
			expr: bson.D{kv("key", mockReducible{val: "field"})},
			expected: bson.D{kv("key", "resolved_field")},
		},
		{
			name: "Nested bson.D and bson.A",
			expr: bson.D{
				kv("nested_a", bson.A{mockReducible{val: "arr_field"}}),
				kv("nested_d", bson.D{kv("inner", mockReducible{val: "dict_field"})}),
			},
			expected: bson.D{
				kv("nested_a", bson.A{"resolved_arr_field"}),
				kv("nested_d", bson.D{kv("inner", "resolved_dict_field")}),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeExpression(tc.expr, resolver)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("NormalizeExpression() = %v, want %v", result, tc.expected)
			}
		})
	}
}
