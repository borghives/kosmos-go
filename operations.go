package kosmos

import (
	"context"
	"log"

	"github.com/borghives/kosmos-go/matter"
	"github.com/borghives/kosmos-go/matter/expression"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Fld(name string) matter.EntityField {
	return matter.EntityField{Name: name}
}

func All[T matter.Detectable]() *matter.EntityDetector[T] {
	return matter.NewEntityDetector[T]()
}

func Detect[T matter.Detectable](filters ...expression.QueryFieldPredicate) *matter.EntityDetector[T] {
	return All[T]().Filter(filters...)
}

func Witness[C matter.Collapsible](ctx context.Context, obj C) error {
	observer := matter.NewEntityObserver[C]()
	return observer.Witness(ctx, obj)
}

func MustHaveObserverClient() *mongo.Client {
	client := matter.SummonMongo(matter.PurposeAffinityObserver).Client()
	if client == nil {
		log.Fatalf("Observer client not initialized")
	}
	return client
}
