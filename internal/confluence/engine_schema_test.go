package confluence

import "testing"

func TestEvaluateUnknownForInvalidInputSchema(t *testing.T) {
	input := testInput("invalid-schema", "receipt.currency", "KRW", "receipt.tax_code", "VAT")
	input.Schema = "gooo/source-input/unknown/v1"
	result, err := EvaluateToDirectory(testSpec(), input, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Decision != Unknown || result.Unknown == nil || len(result.Unknown.BlockedBy) != 1 || result.Unknown.BlockedBy[0] != "input:schema" {
		t.Fatalf("got decision=%q unknown=%+v", result.Decision, result.Unknown)
	}
}
