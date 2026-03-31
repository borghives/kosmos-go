package kosmos_test

import (
	"fmt"
	"testing"

	"github.com/borghives/kosmos-go"
	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TestModel struct {
	kosmos.BaseModel `bson:",inline" kdb:"test_db" kcol:"test_coll"`
	Name             string `bson:"name"`
}

// Ensure TestModel (value) and *TestModel both satisfy Observable
var _ kosmos.Observable = TestModel{}
var _ kosmos.Observable = (*TestModel)(nil)

func TestFilter(t *testing.T) {
	id, _ := bson.ObjectIDFromHex("69cbe858fae0ee418635e8ec")

	// Create a filter matching the id
	record := kosmos.Filter[TestModel](
		model.Fld("_id").Eq(id),
	)
	fmt.Println(record.PipelineJSON())
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
