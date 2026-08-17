package semantic_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/semantic"
)

func TestFactJSONTags(t *testing.T) {
	fact := semantic.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	data, err := json.Marshal(fact)
	if err != nil {
		t.Fatalf(
			"json.Marshal() error = %v",
			err,
		)
	}

	got := string(data)

	for _, want := range []string{
		`"id":"fact-001"`,
		`"statement":"CPU usage is high"`,
		`"confidence":0.98`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf(
				"json = %s, want substring %s",
				got,
				want,
			)
		}
	}

	for _, unwanted := range []string{
		`"metadata"`,
		`"validity"`,
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf(
				"json = %s, unexpectedly contains %s",
				got,
				unwanted,
			)
		}
	}
}

func TestDecisionJSONOmitsEmptyRationale(t *testing.T) {
	decision := semantic.Decision{
		ID:      "decision-001",
		Outcome: "Scale the service",
	}

	data, err := json.Marshal(decision)
	if err != nil {
		t.Fatalf(
			"json.Marshal() error = %v",
			err,
		)
	}

	got := string(data)

	if strings.Contains(got, `"rationale"`) {
		t.Errorf(
			"json = %s, unexpectedly contains rationale",
			got,
		)
	}

	if !strings.Contains(got, `"outcome":"Scale the service"`) {
		t.Errorf(
			"json = %s, want outcome field",
			got,
		)
	}
}

func TestValidityJSONRoundTrip(t *testing.T) {
	validFrom := time.Date(
		2026, time.August, 17, 10, 0, 0, 0, time.UTC,
	)

	fact := semantic.Fact{
		ID:        "fact-001",
		Statement: "The subscription is expired",
		Validity: semantic.Validity{
			ValidFrom: &validFrom,
		},
	}

	data, err := json.Marshal(fact)
	if err != nil {
		t.Fatalf(
			"json.Marshal() error = %v",
			err,
		)
	}

	var decoded semantic.Fact

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf(
			"json.Unmarshal() error = %v",
			err,
		)
	}

	if decoded.Validity.ValidFrom == nil {
		t.Fatal("decoded.Validity.ValidFrom is nil")
	}

	if !decoded.Validity.ValidFrom.Equal(validFrom) {
		t.Errorf(
			"decoded.Validity.ValidFrom = %v, want %v",
			decoded.Validity.ValidFrom,
			validFrom,
		)
	}
}

func TestFactJSONOmitsZeroValidity(t *testing.T) {
	fact := semantic.Fact{
		ID:        "fact-001",
		Statement: "CPU usage is high",
	}

	data, err := json.Marshal(fact)
	if err != nil {
		t.Fatalf(
			"json.Marshal() error = %v",
			err,
		)
	}

	if strings.Contains(string(data), `"validity"`) {
		t.Errorf(
			"json = %s, want zero Validity omitted",
			string(data),
		)
	}
}
