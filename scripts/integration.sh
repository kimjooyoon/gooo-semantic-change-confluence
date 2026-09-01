#!/usr/bin/env bash
set -euo pipefail

workspace=$1
ci_dir=$2
binary=$3
verifier=$4
output_dir="$ci_dir/integration"
mkdir -p "$output_dir/generated"

"$binary" -spec "$workspace/.gooo/confluence.gooo" -input "$workspace/examples/parallel-receipt-changes.gooo" -output "$output_dir/run"
"$verifier" -result "$output_dir/run/result.json" -root "$output_dir/run"
jq -e '.case_id == "normal-independent" and .decision == "CLOSED" and .semantic_verdict == "CONFLUENT" and .merge_authority == 0' "$output_dir/run/result.json" >/dev/null

cp "$output_dir/run/A_then_B/generated.go" "$output_dir/generated/generated.go"
printf '%s\n' 'module gooo-integration-runner' > "$output_dir/go.mod"
mkdir -p "$output_dir/cmd"
printf '%s\n' 'package main' '' 'import (' '  "fmt"' '  "os"' '  "gooo-integration-runner/generated"' ')' '' 'func main() {' '  if generated.Fields["receipt.currency"] != "KRW" || generated.Fields["receipt.tax_code"] != "VAT" {' '    fmt.Fprintln(os.Stderr, "generated artifact did not execute with the expected semantic state")' '    os.Exit(1)' '  }' '  fmt.Println("generated artifact execution verified")' '}' > "$output_dir/cmd/main.go"
(cd "$output_dir" && go run ./cmd)

jq -n \
  --arg scenario "parallel-receipt-changes" \
  --arg source_id "receipt-demo" \
  --arg contract "receipt/v1" \
  --arg toolchain "go1.27.0" \
  --arg runner "gooo-confluence-ci" \
  --arg result "integration/run/result.json" \
  --arg generated "integration/generated/generated.go" \
  '{schema:"gooo/integration-evidence/v1",scenario:$scenario,source_id:$source_id,contract:$contract,toolchain:$toolchain,runner:$runner,execution:"generated artifact executed and verifier passed",result_file:$result,generated_artifact:$generated,merge_authority:0}' > "$ci_dir/integration-report.json"
cp "$output_dir/run/report.md" "$ci_dir/integration-report.md"
