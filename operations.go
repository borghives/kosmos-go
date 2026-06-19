package kosmos

import (
	"context"
	"log"

	"git.mypierian.com/borghives/kosmos-go/matter"
	"git.mypierian.com/borghives/kosmos-go/meta/expression"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Fld(name string) matter.EntityField {
	return matter.EntityField{Name: name}
}

func All[T matter.Detectable]() *matter.Detector[T] {
	return matter.NewDetector[T]()
}

func Detect[T matter.Detectable](filters ...expression.QueryFieldPredicate) *matter.Detector[T] {
	return All[T]().Filter(filters...)
}

func ProjectInto[T matter.Detectable](fields ...expression.FieldSpecification) *matter.Projector[T] {
	return matter.NewProjector[T](fields...)
}

func Record[C matter.Collapsible](ctx context.Context, obj C) error {
	observer := matter.NewObserver[C]()
	return observer.Record(ctx, obj)
}

func MustHaveObserverClient() *mongo.Client {
	client := matter.SummonMongo(matter.PurposeAffinityObserver).Client()
	if client == nil {
		log.Fatalf("Observer client not initialized")
	}
	return client
}
