package observation

import (
	"github.com/borghives/kosmos-go/model"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Entity struct {
	EntityMeta model.Metadata
}

func (e *Entity) DataCollection() *mongo.Collection {
	return SummonMongo(PurposeAffinityObserver).
		BranchDatabase(e.EntityMeta.BranchName).
		Collection(e.EntityMeta.DataName)
}
