package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type Person struct {
	Name    string `properties:"name"`
	Address string `properties:"address,omitempty"`
	Age     int    `properties:"age"`
	Married bool   `properties:"married"`
}

func Serialize(person Person) string {
	v := reflect.ValueOf(person)
	t := reflect.TypeOf(person)
	numField := t.NumField()

	var (
		lines = make([]string, 0, numField)
	)

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("properties")
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")

		omitEmpty := false
		for _, opt := range parts[1:] {
			if opt == "omitempty" {
				omitEmpty = true
				break
			}
		}

		fv := v.Field(i)
		if omitEmpty && fv.IsZero() {
			continue
		}

		var valueStr string
		switch fv.Kind() {
		case reflect.String:
			valueStr = fv.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueStr = fmt.Sprintf("%d", fv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			valueStr = fmt.Sprintf("%d", fv.Uint())
		case reflect.Bool:
			valueStr = fmt.Sprintf("%t", fv.Bool())
		case reflect.Float32, reflect.Float64:
			valueStr = fmt.Sprintf("%v", fv.Float())
		default:
			valueStr = fmt.Sprintf("%v", fv.Interface())
		}

		lines = append(lines, parts[0]+"="+valueStr)
	}

	return strings.Join(lines, "\n")
}

func TestSerialization(t *testing.T) {
	tests := map[string]struct {
		person Person
		result string
	}{
		"test case with empty fields": {
			result: "name=\nage=0\nmarried=false",
		},
		"test case with fields": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
			},
			result: "name=John Doe\nage=30\nmarried=true",
		},
		"test case with omitempty field": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
				Address: "Paris",
			},
			result: "name=John Doe\naddress=Paris\nage=30\nmarried=true",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Serialize(test.person)
			assert.Equal(t, test.result, result)
		})
	}
}
