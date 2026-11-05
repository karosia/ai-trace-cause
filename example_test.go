package aitracecause_test

import (
	"context"
	"fmt"
	"log"

	aitracecause "github.com/karosia/ai-trace-cause"
)

// An agent loop recorded step by step: the first model call queries
// metrics, and the tool result drives the second call's decision. The
// explanation of the final action reaches back through the tool call.
func ExampleRun() {
	ctx := context.Background()

	trace, err := aitracecause.New()
	if err != nil {
		log.Fatal(err)
	}

	run := trace.Run("incident-4521")

	step := run.Step("llm-call-1")
	alert := step.Observation(aitracecause.Observation{Name: "pager_alert", Value: "p99 high"})
	symptom := step.Fact(aitracecause.Fact{Statement: "checkout is slow", Confidence: 0.9}, alert)
	plan := step.Decision(aitracecause.Decision{Outcome: "investigate checkout"}, symptom)
	query := step.Action(aitracecause.Action{Name: "prometheus_query"}, plan)

	if _, err := step.Commit(ctx); err != nil {
		log.Fatal(err)
	}

	step = run.Step("llm-call-2")
	cpu := step.Observation(aitracecause.Observation{Name: "cpu_usage", Value: 97}, query)
	cause := step.Fact(aitracecause.Fact{Statement: "checkout is CPU-bound", Confidence: 0.8}, cpu)
	scale := step.Decision(aitracecause.Decision{Outcome: "scale checkout"}, cause, plan)
	act := step.Action(aitracecause.Action{Name: "scale_service"}, scale)

	if _, err := step.Commit(ctx); err != nil {
		log.Fatal(err)
	}

	explanation, err := trace.Explain(ctx, act, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(explanation)
	// Output:
	// Action "scale_service" was caused by Decision "scale checkout".
	// Decision "scale checkout" refines Decision "investigate checkout".
	// Decision "scale checkout" was based on Fact "checkout is CPU-bound".
	// Decision "investigate checkout" was based on Fact "checkout is slow".
	// Fact "checkout is CPU-bound" was supported by Observation "cpu_usage" (value=97).
	// Fact "checkout is slow" was supported by Observation "pager_alert" (value=p99 high).
	// Observation "cpu_usage" (value=97) resulted from Action "prometheus_query".
	// Action "prometheus_query" was caused by Decision "investigate checkout".
}
