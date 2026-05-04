package expression

import (
	"github.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Scope []QueryFieldPredicate

func CreateScope(filters ...QueryFieldPredicate) Scope {
	return Scope(filters)
}

func (s Scope) Express(objMeta meta.Metadata) bson.D {
	return NormalizeExpression(AndField(s...), objMeta.ResolveAlias).(bson.D)
}
