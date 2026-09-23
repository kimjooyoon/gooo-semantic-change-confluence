package confluence

import "testing"

func TestOperationConflictRejectsDifferentValuesFromSameAuthority(t *testing.T) {
	left := Operation{ID: "left", Authority: "same-owner", Patches: []Patch{{Path: "receipt.currency", Value: "KRW"}}}
	right := Operation{ID: "right", Authority: "same-owner", Patches: []Patch{{Path: "receipt.currency", Value: "USD"}}}

	if reason, conflict := operationConflict(Spec{}, left, right); !conflict || reason == "" {
		t.Fatalf("operationConflict() = %q, %v; want a semantic value conflict", reason, conflict)
	}
}

func TestOperationConflictAllowsIdempotentSameValue(t *testing.T) {
	left := Operation{ID: "left", Authority: "same-owner", Patches: []Patch{{Path: "receipt.currency", Value: "KRW"}}}
	right := Operation{ID: "right", Authority: "same-owner", Patches: []Patch{{Path: "receipt.currency", Value: "KRW"}}}

	if reason, conflict := operationConflict(Spec{}, left, right); conflict {
		t.Fatalf("operationConflict() = %q, %v; want idempotent writes to remain compatible", reason, conflict)
	}
}
