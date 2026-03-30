package entitymodel

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityRecord[T EntityModel] struct {
	Stages Aggregation
	Type   EntityType
}

func (r *EntityRecord[T]) Filter(filter QueryPredicate) *EntityRecord[T] {

	r.Stages = r.Stages.Match(r.Type.NormalizeExpression(filter).(bson.D))
	return r
}

func (r *EntityRecord[T]) Sort(field string, descending bool) *EntityRecord[T] {
	order := 1
	if descending {
		order = -1
	}
	r.Stages = r.Stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *EntityRecord[T]) PullOne() *T {
	results, err := r.pullPipeline(Aggregation{}.Limit(1))
	if err != nil {
		return nil
	}
	if len(results) == 0 {
		return nil
	}
	return results[0]
}

func (r *EntityRecord[T]) PullAll() []*T {
	results, err := r.pullPipeline(Aggregation{})
	if err != nil {
		return nil
	}
	return results
}

func (r *EntityRecord[T]) dataCollection() *mongo.Collection {
	return r.Type.Collection
}

func (r *EntityRecord[T]) pullPipeline(postStages Aggregation) ([]*T, error) {
	collection := r.dataCollection()

	pipeline := r.Stages.AppendFrom(postStages).Pipeline()

	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []*T
	cursor.All(context.Background(), &results)
	return results, nil
}
