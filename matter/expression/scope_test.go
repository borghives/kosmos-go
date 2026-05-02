package expression

import (
	"reflect"
	"testing"

	"github.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestScope_Express(t *testing.T) {
	q1 := QueryFieldPredicate{FieldName: FieldName{Name: "alias_me"}, Query: Eq(1)}
	q2 := QueryFieldPredicate{FieldName: FieldName{Name: "normal_field"}, Query: Gt(2)}
	
	s := CreateScope(q1, q2)
	
	m := meta.Metadata{
		FieldMap: map[string]string{"alias_me": "resolved_alias"},
	}
	
	result := s.Express(m)

	expected := bson.D{
		kv("$and", bson.A{
			bson.D{kv("resolved_alias", bson.D{kv("$eq", 1)})},
			bson.D{kv("normal_field", bson.D{kv("$gt", 2)})},
		}),
	}
	
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Scope.Express() = %v, want %v", result, expected)
	}
}
