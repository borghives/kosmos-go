package matter

import (
	"git.mypierian.com/borghives/kosmos-go/meta"
	"git.mypierian.com/borghives/kosmos-go/meta/expression"
)

type Projector[T Detectable] struct {
	FieldsSpecs []expression.FieldSpecification
}

func NewProjector[T Detectable](fields ...expression.FieldSpecification) *Projector[T] {
	return &Projector[T]{
		FieldsSpecs: fields,
	}
}

func (a Projector[T]) From(detector DetectionOp) *Detector[T] {
	var template T
	projectionEntityMeta := meta.GetMetadata(template)
	project := expression.ReduceFieldSpecification(projectionEntityMeta.ResolveName, a.FieldsSpecs...)

	hasID := false
	for _, spec := range project {
		if spec.Key == "_id" {
			hasID = true
			break
		}
	}

	if !hasID {
		project = append(project, kv("_id", 0))
	}

	retval := &Detector[T]{
		Dataverse: detector.OpContext(),
		stages:    detector.OpStages().Project(project),
	}

	return retval
}
