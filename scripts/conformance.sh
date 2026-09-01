#!/usr/bin/env bash
set -euo pipefail

workspace=$1
ci_dir=$2
binary=$3
verifier=$4
run_root="$ci_dir/runs"
summary_dir="$ci_dir/conformance/cases"
mkdir -p "$run_root" "$summary_dir"

for fixture in "$workspace"/fixtures/*.gooo; do
  case_id=$(jq -r '.case_id' "$fixture")
  expected=$(jq -r --arg id "$case_id" '.canonical_cases[] | select(.id == $id) | .expected' "$workspace/.gooo/confluence.gooo")
  test -n "$expected" && test "$expected" != "null"
  output_dir="$run_root/$case_id"
  mkdir -p "$output_dir"
  "$binary" -spec "$workspace/.gooo/confluence.gooo" -input "$fixture" -output "$output_dir"
  "$verifier" -result "$output_dir/result.json" -root "$output_dir"
  actual=$(jq -r '.decision' "$output_dir/result.json")
  test "$actual" = "$expected"
  jq -n --arg case_id "$case_id" --arg kind "$(jq -r --arg id "$case_id" '.canonical_cases[] | select(.id == $id) | .kind' "$workspace/.gooo/confluence.gooo")" \
    --arg expected "$expected" --arg actual "$actual" \
    --arg reason "$(jq -r 'if .decision == "UNKNOWN" then .unknown.reason elif .decision == "REFUTED" then .counterexample.reason else "both orders have equal canonical IR, generated artifact, and provenance" end' "$output_dir/result.json")" \
    '{case_id:$case_id,kind:$kind,expected:$expected,actual:$actual,reason:$reason}' > "$summary_dir/$case_id.json"
done

jq -s '
  def count($value): map(select(.actual == $value)) | length;
  {
    schema:"gooo/confluence-conformance/v1",
    cases: sort_by(.case_id),
    counts:{normal:map(select(.kind == "normal"))|length,unknown:count("UNKNOWN"),refuted:count("REFUTED"),closed:count("CLOSED")},
    tests:{total:length,selected:length,executed:length,reused:0,failed:map(select(.actual != .expected))|length,unknown:count("UNKNOWN")}
  }
' "$summary_dir"/*.json > "$ci_dir/conformance-report.json"

jq -e '.counts.normal == 3 and .counts.unknown == 3 and .counts.refuted == 3 and .counts.closed == 3 and .tests.total == 9 and .tests.selected == 9 and .tests.executed == 9 and .tests.reused == 0 and .tests.failed == 0 and .tests.unknown == 3' "$ci_dir/conformance-report.json" >/dev/null

{
  echo '# Canonical confluence report'
  echo
  echo 'The denominator is fixed at nine canonical cases: three normal, three UNKNOWN, and three REFUTED.'
  echo
  echo '| Case | Kind | Expected | Actual | Reason |'
  echo '|---|---|---|---|---|'
  jq -r '.cases[] | "| `\(.case_id)` | \(.kind) | `\(.expected)` | `\(.actual)` | \(.reason) |"' "$ci_dir/conformance-report.json"
} > "$ci_dir/conformance-report.md"
