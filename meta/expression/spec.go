package expression

import "go.mongodb.org/mongo-driver/v2/bson"

type FieldSpecification struct {
	FieldName  FieldName
	Expression Base
}

func (f FieldSpecification) ToRepr() any {
	return bson.D{kv(f.FieldName.Name, f.Expression.ToRepr())}
}

func (f FieldSpecification) Reduce(resolver NameResolver) any {
	return bson.D{kv(resolver(f.FieldName.Name), NormalizeExpression(f.Expression, resolver))}
}

func ReduceFieldSpecification(resolver NameResolver, fields ...FieldSpecification) bson.D {
	projectDoc := bson.D{}
	for _, field := range fields {
		spec := field.Reduce(resolver).(bson.D)
		projectDoc = append(projectDoc, spec...)
	}
	return projectDoc
}
