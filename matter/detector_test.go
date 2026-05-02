package matter

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockDetectable struct {
	BaseModel struct{}    `kosmos:"test"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UpdatedAt time.Time     `bson:"updated_time,omitempty"`
	Name      string        `bson:"name,omitempty"`
	Age       int           `bson:"age,omitempty"`
}

func (m mockDetectable) GetID() bson.ObjectID {
	return m.ID
}

func (m mockDetectable) HasID() bool {
	return !m.ID.IsZero()
}

func (m mockDetectable) LastObserved() time.Time {
	return m.UpdatedAt
}

func TestDetector_FilterEither(t *testing.T) {
	nameField := EntityField{Name: "name"}.Str()
	ageField := EntityField{Name: "age"}

	t.Run("empty filters", func(t *testing.T) {
		detector := NewDetector[mockDetectable]()
		res := detector.FilterEither()
		
		pipeline := res.stages.Pipeline()
		if len(pipeline) != 0 {
			t.Errorf("expected empty pipeline, got %v", pipeline)
		}
	})

	t.Run("single filter", func(t *testing.T) {
		detector := NewDetector[mockDetectable]()
		res := detector.FilterEither(nameField.Eq("alice"))
		
		pipeline := res.stages.Pipeline()
		if len(pipeline) != 1 {
			t.Fatalf("expected pipeline length 1, got %d", len(pipeline))
		}

		expectedMatch := bson.D{kv("$match", bson.D{kv("name", bson.D{kv("$eq", "alice")})})}
		if !reflect.DeepEqual(pipeline[0], expectedMatch) {
			t.Errorf("expected %v, got %v", expectedMatch, pipeline[0])
		}
	})

	t.Run("multiple filters", func(t *testing.T) {
		detector := NewDetector[mockDetectable]()
		res := detector.FilterEither(
			nameField.Eq("alice"),
			ageField.Gt(30),
		)
		
		pipeline := res.stages.Pipeline()
		if len(pipeline) != 1 {
			t.Fatalf("expected pipeline length 1, got %d", len(pipeline))
		}

		expectedOr := bson.D{kv("$or", bson.A{
			bson.D{kv("name", bson.D{kv("$eq", "alice")})},
			bson.D{kv("age", bson.D{kv("$gt", 30)})},
		})}
		expectedMatch := bson.D{kv("$match", expectedOr)}

		if !reflect.DeepEqual(pipeline[0], expectedMatch) {
			t.Errorf("expected %v, got %v", expectedMatch, pipeline[0])
		}
	})
}
