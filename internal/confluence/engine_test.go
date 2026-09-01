package confluence

import (
	"testing"
)

func testSpec() Spec {
	return Spec{
		SemanticGraph: SemanticGraph{Nodes: []GraphNode{{ID: "source", Kind: "source"}, {ID: "semantic-ir", Kind: "ir"}}},
		ChangeOperations: []Operation{{ID: "change-a"}, {ID: "change-b"}},
		ApplicationOrders: map[string][]string{"A_then_B": {"change-a", "change-b"}, "B_then_A": {"change-b", "change-a"}},
		Authorities: Authorities{MergeAuthority: 0, RuntimeWrite: 0},
		Guardrails: Guardrails{RequiredInputPaths: []string{"receipt.id"}, RequiredOperationFields: []string{"scope", "authority", "patches"}, ForbiddenPathPrefixes: []string{"receipt.private"}},
	}
}

func testInput(caseID string, pathA, valueA, pathB, valueB string) SourceInput {
	return SourceInput{
		CaseID: caseID, SourceID: "test-source", Contract: "test/v1", Toolchain: "go1.27.0", Runner: "test-runner",
		Baseline: map[string]string{"receipt.id": "test"},
		Operations: []Operation{
			{ID: "change-a", Authority: "source-owner-a", Scope: []string{pathA}, Patches: []Patch{{Path: pathA, Value: valueA}}},
			{ID: "change-b", Authority: "source-owner-b", Scope: []string{pathB}, Patches: []Patch{{Path: pathB, Value: valueB}}},
		},
	}
}

func TestEvaluateClosedForDisjointChanges(t *testing.T) {
	result, err := EvaluateToDirectory(testSpec(), testInput("closed", "receipt.currency", "KRW", "receipt.tax_code", "VAT"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != Closed || result.SemanticVerdict != "CONFLUENT" {
		t.Fatalf("got decision=%q verdict=%q", result.Decision, result.SemanticVerdict)
	}
	if result.Orders["A_then_B"].SemanticIRDigest != result.Orders["B_then_A"].SemanticIRDigest {
		t.Fatal("canonical IR differs for disjoint changes")
	}
}

func TestEvaluateUnknownCarriesFrontier(t *testing.T) {
	input := testInput("unknown", "receipt.currency", "KRW", "receipt.tax_code", "VAT")
	input.Operations[1].DependsOn = []string{"missing-operation"}
	result, err := EvaluateToDirectory(testSpec(), input, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != Unknown || result.Unknown == nil || len(result.Unknown.BlockedBy) != 1 {
		t.Fatalf("got decision=%q unknown=%+v", result.Decision, result.Unknown)
	}
}

func TestEvaluateRefutedForOrderDependentChanges(t *testing.T) {
	result, err := EvaluateToDirectory(testSpec(), testInput("refuted", "receipt.currency", "KRW", "receipt.currency", "USD"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != Refuted || result.Counterexample == nil || result.Counterexample.Path != "receipt.currency" {
		t.Fatalf("got decision=%q counterexample=%+v", result.Decision, result.Counterexample)
	}
}
