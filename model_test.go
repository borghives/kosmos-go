package kosmos_test

import (
	"testing"

	"github.com/borghives/kosmos-go"
	km "github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TestModel struct {
	km.BaseModel `bson:",inline" kosmos:"test_coll"`
	Name         string `bson:"name"`
}

// Ensure TestModel (value) and *TestModel both satisfy Observable
var _ model.Observable = TestModel{}
var _ model.Observable = (*TestModel)(nil)

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

	if record.Type.CollectionName != "test_coll" {
		t.Errorf("expected collection name 'test_coll', got '%s'", record.Type.CollectionName)
	}

	obj, err := record.PullOne()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
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

func TestFilterPredicate(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("69cbe858fae0ee418635e8ec")

	// Create a filter matching the id
	record := kosmos.Filter[TestModel](
		km.Fld("_id").ID().Eq(id),
		km.Fld("name").Str().Eq("MAGIC"),
	)
	if record == nil {
		t.Fatalf("expected record to not be nil")
	}

	if record.Type.CollectionName != "test_coll" {
		t.Errorf("expected collection name 'test_coll', got '%s'", record.Type.CollectionName)
	}

	obj, err := record.PullOne()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
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

func TestFilterIn(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("69cbe858fae0ee418635e8ec")
	id2, _ := bson.ObjectIDFromHex("69cbe858fae0ee418635e8ed")

	ids := []bson.ObjectID{id, id2}

	// Create a filter matching the id
	record := kosmos.Filter[TestModel](
		km.Fld("_id").ID().In(ids...),
	)
	if record == nil {
		t.Fatalf("expected record to not be nil")
	}

	if record.Type.CollectionName != "test_coll" {
		t.Errorf("expected collection name 'test_coll', got '%s'", record.Type.CollectionName)
	}

	obj, err := record.PullOne()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
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
	detector := kosmos.Filter[*TestModel](
		km.Fld("_id").Eq(bson.NewObjectID()),
	)
	if detector.Type.CollectionName != "test_coll" {
		t.Errorf("expected test_coll, got %s", detector.Type.CollectionName)
	}
}

func TestNormalizeDocument(t *testing.T) {
	meta := model.GetMetadata(TestModel{})

	// Create an expression that should trigger FieldName rewrite.
	exprID := km.Fld("ID").Eq(123)
	exprName := km.Fld("Name").Eq("MAGIC")

	// Test that 'ID' and 'Name' are mapped to '_id' and 'name' ONLY when used via Expressions.
	// Raw string keys should NOT be mapped.
	doc := bson.D{
		{Key: "ID", Value: 1},
		{Key: "Name", Value: "Raw"},
		{Key: "query", Value: bson.A{exprID, exprName}},
	}

	norm := expression.NormalizeDocument(doc, meta.ResolveAlias)

	// Check unmapped keys
	if norm[0].Key != "ID" {
		t.Errorf("expected raw 'ID' key to be unchanged, got %s", norm[0].Key)
	}
	if norm[1].Key != "Name" {
		t.Errorf("expected raw 'Name' key to be unchanged, got %s", norm[1].Key)
	}

	// Unpack the array to find mapped expressions
	arrValue := norm[2].Value.(bson.A)
	if len(arrValue) != 2 {
		t.Fatalf("expected 2 array elements, got %d", len(arrValue))
	}

	idDoc := arrValue[0].(bson.D)
	if len(idDoc) == 0 || idDoc[0].Key != "_id" {
		t.Errorf("expected 'ID' to be normalized to '_id', got %s", idDoc[0].Key)
	}

	nameDoc := arrValue[1].(bson.D)
	if len(nameDoc) == 0 || nameDoc[0].Key != "name" {
		t.Errorf("expected 'Name' to be normalized to 'name', got %s", nameDoc[0].Key)
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
