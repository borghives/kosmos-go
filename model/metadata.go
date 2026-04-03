package model

import (
	"log"
	"reflect"
	"slices"
	"strings"
)

type Metadata struct {
	DataName   string
	BranchName string
	FieldMap   map[string]string
}

func (e *Metadata) ResolveAlias(name string) string {
	if e.FieldMap != nil {
		if mapped, ok := e.FieldMap[name]; ok {
			return mapped
		}
	}
	return name
}

func GetMetadata(obj any) Metadata {
	t := reflect.TypeOf(obj)
	if t == nil {
		log.Fatal("model.GetMetadata: cannot extract metadata from a nil interface")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	field, found := t.FieldByName("BaseModel")
	if !found {
		log.Fatal("model.GetMetadata: BaseModel not found")
	}

	fieldMap := make(map[string]string)
	populateFieldMap(t, fieldMap)

	var dataName string
	var branchName string
	modelKosmosString := field.Tag.Get("kosmos")
	kosmosParts := strings.Split(modelKosmosString, ",")
	if len(kosmosParts) > 0 {
		// handle dataverse and collection the first part of the kosmos string
		dataParts := strings.Split(kosmosParts[0], ">")

		if len(dataParts) == 1 {
			dataName = dataParts[0]
		} else if len(dataParts) == 2 {
			branchName = dataParts[0]
			dataName = dataParts[1]
		} else {
			log.Fatalf("model.GetMetadata: invalid kosmos data tag format: %s", kosmosParts[0])
		}
	}

	return Metadata{
		DataName:   dataName,
		BranchName: branchName,
		FieldMap:   fieldMap,
	}
}

func populateFieldMap(t reflect.Type, m map[string]string) {
	if t.Kind() != reflect.Struct {
		return
	}
	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		bsonTag := field.Tag.Get("bson")
		if bsonTag == "-" {
			continue
		}

		bsonName := field.Name
		inline := false
		if bsonTag != "" {
			parts := strings.Split(bsonTag, ",")
			if parts[0] != "" {
				bsonName = parts[0]
			}
			inline = slices.Contains(parts[1:], "inline")
		}

		m[field.Name] = bsonName

		if inline || field.Anonymous {
			fieldType := field.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			if fieldType.Kind() == reflect.Struct {
				populateFieldMap(fieldType, m)
			}
		}
	}
}
