package expression

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestQueryFieldPredicate_Empty(t *testing.T) {
	// Empty case
	var empty QueryFieldPredicate
	if !empty.Empty() {
		t.Errorf("expected empty QueryFieldPredicate to be empty")
	}

	// Non-empty case (safe from panics even with uncomparable slices)
	nonEmpty := QueryFieldPredicate{
		FieldName: FieldName{Name: "test"},
		Query:     In(bson.A{1, 2, 3}),
	}
	if nonEmpty.Empty() {
		t.Errorf("expected non-empty QueryFieldPredicate not to be empty")
	}
}

func TestQueryOperators(t *testing.T) {
	tests := []struct {
		name     string
		op       QueryOp
		expected any
	}{
		{"Eq", Eq(1), bson.D{kv("$eq", 1)}},
		{"Ne", Ne(2), bson.D{kv("$ne", 2)}},
		{"Gt", Gt(3), bson.D{kv("$gt", 3)}},
		{"Gte", Gte(4), bson.D{kv("$gte", 4)}},
		{"Lt", Lt(5), bson.D{kv("$lt", 5)}},
		{"Lte", Lte(6), bson.D{kv("$lte", 6)}},
		{"In", In(bson.A{1, 2}), bson.D{kv("$in", bson.A{1, 2})}},
		{"Nin", Nin(bson.A{3, 4}), bson.D{kv("$nin", bson.A{3, 4})}},
		{"And", And(bson.A{bson.D{kv("a", 1)}}), bson.D{kv("$and", bson.A{bson.D{kv("a", 1)}})}},
		{"Or", Or(bson.A{bson.D{kv("b", 2)}}), bson.D{kv("$or", bson.A{bson.D{kv("b", 2)}})}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.op.ToRepr()
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("%s() = %v, want %v", tc.name, result, tc.expected)
			}
		})
	}
}

func TestQueryFieldPredicate(t *testing.T) {
	resolver := func(s string) string {
		return "resolved_" + s
	}

	q := QueryFieldPredicate{
		FieldName: FieldName{Name: "age"},
		Query:     Gt(30),
	}

	expectedRepr := bson.D{kv("age", bson.D{kv("$gt", 30)})}
	if !reflect.DeepEqual(q.ToRepr(), expectedRepr) {
		t.Errorf("ToRepr() = %v, want %v", q.ToRepr(), expectedRepr)
	}

	expectedReduce := bson.D{kv("resolved_age", bson.D{kv("$gt", 30)})}
	if !reflect.DeepEqual(q.Reduce(resolver), expectedReduce) {
		t.Errorf("Reduce() = %v, want %v", q.Reduce(resolver), expectedReduce)
	}

	// Empty predicate test for ToRepr and Reduce
	var empty QueryFieldPredicate
	if !reflect.DeepEqual(empty.ToRepr(), bson.D{}) {
		t.Errorf("Empty ToRepr() should be empty bson.D")
	}
	if !reflect.DeepEqual(empty.Reduce(resolver), bson.D{}) {
		t.Errorf("Empty Reduce() should be empty bson.D")
	}
}

func TestOrFieldAndField(t *testing.T) {
	q1 := QueryFieldPredicate{FieldName: FieldName{Name: "a"}, Query: Eq(1)}
	q2 := QueryFieldPredicate{FieldName: FieldName{Name: "b"}, Query: Eq(2)}
	var empty QueryFieldPredicate

	t.Run("OrField empty", func(t *testing.T) {
		orEmpty := OrField()
		if !reflect.DeepEqual(orEmpty.ToRepr(), bson.D{}) {
			t.Errorf("expected empty bson.D")
		}
	})

	t.Run("OrField single", func(t *testing.T) {
		orSingle := OrField(q1)
		if !reflect.DeepEqual(orSingle.ToRepr(), q1) {
			t.Errorf("expected single element returned directly")
		}
	})

	t.Run("OrField multiple", func(t *testing.T) {
		orMulti := OrField(q1, q2)
		expected := Or(bson.A{q1, q2})
		if !reflect.DeepEqual(orMulti.ToRepr(), expected) {
			t.Errorf("expected Or with array of queries")
		}
	})

	t.Run("AndField empty", func(t *testing.T) {
		andEmpty := AndField()
		if !reflect.DeepEqual(andEmpty.ToRepr(), bson.D{}) {
			t.Errorf("expected empty bson.D")
		}
	})

	t.Run("AndField single", func(t *testing.T) {
		andSingle := AndField(q1)
		if !reflect.DeepEqual(andSingle.ToRepr(), q1) {
			t.Errorf("expected single element returned directly")
		}
	})

	t.Run("AndField multiple", func(t *testing.T) {
		andMulti := AndField(q1, q2)
		expected := And(bson.A{q1, q2})
		if !reflect.DeepEqual(andMulti.ToRepr(), expected) {
			t.Errorf("expected And with array of queries")
		}
	})

	t.Run("ToBSONArray with empty elements", func(t *testing.T) {
		arr := ToBSONArray(q1, empty, q2)
		expected := bson.A{q1, q2}
		if !reflect.DeepEqual(arr, expected) {
			t.Errorf("ToBSONArray() = %v, want %v", arr, expected)
		}
	})
}

func TestScope(t *testing.T) {
	// Simple test for Scope since it's just a slice alias for now
	// If scope.go has actual logic, test it here
	s := CreateScope(QueryFieldPredicate{FieldName: FieldName{Name: "a"}, Query: Eq(1)})
	if len(s) != 1 {
		t.Errorf("expected scope length 1")
	}
}
