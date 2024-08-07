package sources

import (
	"encoding/json"
	"testing"
)

func TestCamelCase(t *testing.T) {
	exampleMap := make(map[string]interface{})

	exampleMap["Name"] = "Dylan"
	exampleMap["Nested"] = map[string]interface{}{
		"NestedKeyName":    "Value",
		"NestedAWSAcronym": "Wow",
		"NestedArray": []map[string]string{
			{
				"FooBar": "Baz",
			},
		},
	}

	i := interface{}(exampleMap)

	camel := CamelCase(i)

	b, err := json.Marshal(camel)
	if err != nil {
		t.Fatalf("error marshalling: %v", err)
	}

	expected := `{"name":"Dylan","nested":{"nestedAWSAcronym":"Wow","nestedArray":[{"fooBar":"Baz"}],"nestedKeyName":"Value"}}`

	if string(b) != expected {
		t.Fatalf("expected %v got %v", expected, string(b))
	}
}

func TestToAttributesCase(t *testing.T) {
	exampleMap := make(map[string]interface{})

	exampleMap["Name"] = "Dylan"
	exampleMap["Removed"] = "goodbye"
	exampleMap["Nested"] = map[string]string{
		"NestedKeyName":    "Value",
		"NestedAWSAcronym": "Wow",
	}
	exampleMap["Nil"] = nil

	i := interface{}(exampleMap)

	attrs, err := ToAttributesCase(i, "removed")

	if err != nil {
		t.Fatal(err)
	}

	if _, err := attrs.Get("nested"); err != nil {
		t.Error("could not find key nested")
	}

	if _, err := attrs.Get("nil"); err == nil {
		t.Error("expected nil attributes to be removed")
	}

	if _, err := attrs.Get("removed"); err == nil {
		t.Error("expected 'removed' to have been removed")
	}
}
