package confluence

import "testing"

func TestEvaluateUnknownForInvalidSemanticContract(t *testing.T) {
	spec := testSpec()
	spec.Schema = "gooo/semantic-change-confluence/unknown/v1"
	result, err := EvaluateToDirectory(spec, testInput("invalid-contract", "receipt.currency", "KRW", "receipt.tax_code", "VAT"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != Unknown || result.Unknown == nil || len(result.Unknown.BlockedBy) != 1 || result.Unknown.BlockedBy[0] != "meta:schema" {
		t.Fatalf("got decision=%q unknown=%+v", result.Decision, result.Unknown)
	}
}
