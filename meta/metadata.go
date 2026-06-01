package meta

import (
	"log"
	"reflect"
	"slices"
	"strings"
)

type MetaState struct {
	DataName   string
	BranchName string
	FieldMap   map[string]string
}

func (e *MetaState) ResolveName(name string) string {
	if e.FieldMap != nil {
		if mapped, ok := e.FieldMap[name]; ok {
			return mapped
		}
	}
	return name
}

func GetMetadata(obj any) MetaState {
	t := reflect.TypeOf(obj)
	if t == nil {
		log.Fatal("meta.GetMetadata: cannot extract metadata from a nil interface")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	fieldMap := make(map[string]string)
	populateFieldMap(t, fieldMap)

	var dataName string
	var branchName string

	field, found := t.FieldByName("BaseModel")
	if found {
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
				log.Fatalf("meta.GetMetadata: invalid kosmos data tag format: %s", kosmosParts[0])
			}
		}
	}

	return MetaState{
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
