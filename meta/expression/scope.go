package expression

import (
	"git.mypierian.com/borghives/kosmos-go/meta"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Scope []QueryFieldPredicate

func CreateScope(filters ...QueryFieldPredicate) Scope {
	return Scope(filters)
}

func (s Scope) Express(objMeta meta.MetaState) bson.D {
	return NormalizeExpression(AndField(s...), objMeta.ResolveName).(bson.D)
}
