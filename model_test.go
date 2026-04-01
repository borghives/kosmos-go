package kosmos_test

import (
	"testing"

	"github.com/borghives/kosmos-go"
	km "github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TestModel struct {
	kosmos.BaseModel `bson:",inline" kdb:"test_db" kcol:"test_coll"`
	Name             string `bson:"name"`
}

// Ensure TestModel (value) and *TestModel both satisfy Observable
var _ km.Observable = TestModel{}
var _ km.Observable = (*TestModel)(nil)

func TestWitness(t *testing.T) {
	m := TestModel{
		Name: "MAGIC_Ed4",
	}
	kosmos.Witness(&m)
}

func TestFilter(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("69cbe858fae0ee418635e8ec")

	// Create a filter matching the id
	record := kosmos.Filter[TestModel](
		km.Fld("_id").Eq(id),
	)
	if record == nil {
		t.Fatalf("expected record to not be nil")
	}

	// Verify that the metadata was extracted properly
	if record.Type.DatabaseName != "test_db" {
		t.Errorf("expected database name 'test_db', got '%s'", record.Type.DatabaseName)
	}
	if record.Type.CollectionName != "test_coll" {
		t.Errorf("expected collection name 'test_coll', got '%s'", record.Type.CollectionName)
	}

	obj := record.PullOne()
	if obj == nil {
		t.Fatalf("expected object to not be nil")
	}
	if obj.ID != id {
		t.Errorf("expected object id '%s', got '%s'", id, obj.ID)
	}
	if obj.Name != "MAGIC" {
		t.Errorf("expected object name 'test', got '%s'", obj.Name)
	}
}

func TestFilterPointer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Filterpanicked with pointer model: %v", r)
		}
	}()
	// Create a filter with pointer T
	record := kosmos.Filter[*TestModel](
		km.Fld("_id").Eq(bson.NewObjectID()),
	)
	if record.Type.DatabaseName != "test_db" {
		t.Errorf("expected test_db, got %s", record.Type.DatabaseName)
	}
}

func TestNormalizeDocument(t *testing.T) {
	meta := km.Metadata{}
	doc := bson.D{
		{Key: "a", Value: bson.D{{Key: "$eq", Value: 1}}},
		{Key: "b", Value: km.Fld("field").Eq(2)},
		{Key: "c", Value: bson.A{1, km.Fld("field2").Eq(3)}},
	}
	norm := meta.NormalizeDocument(doc)

	// Just check if normalization succeeded without panicking and basic structure maintained.
	if len(norm) != 3 {
		t.Errorf("expected length 3, got %d", len(norm))
	}
}

func TestBaseModelCollapseID(t *testing.T) {
	m := kosmos.BaseModel{}
	if m.IsEntangled() {
		t.Error("expected new model to not be entangled")
	}

	id := m.CollapseID()
	if id.IsZero() {
		t.Error("expected non-zero id after collapse")
	}
	if !m.IsEntangled() {
		t.Error("expected model to be entangled after collapse")
	}
	if m.ID != id {
		t.Errorf("expected m.ID to be %v, got %v", id, m.ID)
	}
	if m.InitialObserved().IsZero() {
		t.Error("expected InitialObserved to be set")
	}
}

func TestFilterOperators(t *testing.T) {
	record := kosmos.Filter[TestModel](km.Fld("age").Gt(18)).Sort("name", false)
	json := record.PipelineJSON()
	if json == "" {
		t.Error("expected valid pipeline json")
	}

	// Just invoke other operators to ensure they build correctly without panic
	km.Fld("age").Gte(18)
	km.Fld("age").Lt(18)
	km.Fld("age").Lte(18)
	km.Fld("age").Ne(18)
	km.Fld("status").In("active", "pending")
	km.Fld("status").Nin("banned")
}
