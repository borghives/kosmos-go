package kosmos

import (
	"context"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityRecord[T Model] struct {
	Type   ModelMetadata
	stages model.Aggregation
}

func (r *EntityRecord[T]) Filter(filter model.QueryPredicate) *EntityRecord[T] {

	r.stages = r.stages.Match(r.Type.NormalizeExpression(filter).(bson.D))
	return r
}

func (r *EntityRecord[T]) Sort(field string, descending bool) *EntityRecord[T] {
	order := 1
	if descending {
		order = -1
	}
	r.stages = r.stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *EntityRecord[T]) PullOne() *T {
	results, err := r.pullPipeline(model.Aggregation{}.Limit(1))
	if err != nil {
		return nil
	}
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

func (r *EntityRecord[T]) PullAll() []*T {
	results, err := r.pullPipeline(model.Aggregation{})
	if err != nil {
		return nil
	}
	return results
}

func (r *EntityRecord[T]) dataCollection() *mongo.Collection {
	observer := observation.SummonMongo(observation.PurposeAffinityObserver)
	return observer.Database(r.Type.DatabaseName).Collection(r.Type.CollectionName)
}

func (r *EntityRecord[T]) pullPipeline(postStages model.Aggregation) ([]*T, error) {
	collection := r.dataCollection()

	pipeline := r.stages.AppendFrom(postStages).Pipeline()

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*T
	cursor.All(context.Background(), &results)
	return results, nil
}
