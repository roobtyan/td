package schema

import "testing"

func TestParseTaskSchema(t *testing.T) {
	sampleResponse := `{
		"title": "Buy milk",
		"notes": "remember lactose free",
		"project": "home",
		"priority": "P2",
		"links": ["https://example.com/item"]
	}`
	if err := ValidateParseTaskJSON(sampleResponse); err != nil {
		t.Fatalf("validate schema: %v", err)
	}
}
