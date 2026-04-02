package observation

import (
	"context"
	"fmt"

	"github.com/borghives/kosmos-go/model"
	"github.com/borghives/kosmos-go/observation/expression"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EntityDetector[T model.Observable] struct {
	Type   model.Metadata
	stages Aggregation
}

func (r *EntityDetector[T]) Filter(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.Type.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.And(exprs), r.Type.ResolveAlias).(bson.D))
	}
	return r
}

func (r *EntityDetector[T]) FilterEither(filters ...expression.QueryFieldPredicate) *EntityDetector[T] {
	if len(filters) == 0 {
		return r
	} else if len(filters) == 1 {
		r.stages = r.stages.Match(expression.NormalizeExpression(filters[0], r.Type.ResolveAlias).(bson.D))
	} else {
		exprs := make(bson.A, len(filters))
		for i, f := range filters {
			exprs[i] = f
		}
		r.stages = r.stages.Match(expression.NormalizeExpression(expression.Or(exprs), r.Type.ResolveAlias).(bson.D))
	}
	return r
}

func (r *EntityDetector[T]) Sort(field string, descending bool) *EntityDetector[T] {
	order := 1
	if descending {
		order = -1
	}
	r.stages = r.stages.Sort(bson.D{kv(field, order)})
	return r
}

func (r *EntityDetector[T]) PullOne() *T {
	results, err := r.pullPipeline(Aggregation{}.Limit(1))
	if err != nil {
		return nil
	}
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

func (r *EntityDetector[T]) PullAll() []T {
	results, err := r.pullPipeline(Aggregation{})
	if err != nil {
		return nil
	}
	return results
}

func (r *EntityDetector[T]) dataCollection() *mongo.Collection {
	observer := SummonMongo(PurposeAffinityObserver)
	return observer.Database(r.Type.DatabaseName).Collection(r.Type.CollectionName)
}

func (r *EntityDetector[T]) PipelineJSON() string {
	return r.stages.JsonString()
}

func (r *EntityDetector[T]) pullPipeline(postStages Aggregation) ([]T, error) {
	collection := r.dataCollection()
	pipeline := r.stages.AppendFrom(postStages).Pipeline()
	fmt.Println(r.stages.AppendFrom(postStages).JsonString())
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []T
	cursor.All(context.Background(), &results)
	return results, nil
}
