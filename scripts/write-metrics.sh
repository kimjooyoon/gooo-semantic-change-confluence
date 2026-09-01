#!/usr/bin/env bash
set -euo pipefail

workspace=$1
ci_dir=$2
output=$3

count_files() {
  find "$workspace" -type f -not -path "$workspace/.git/*" -not -path "$workspace/.ci/*" -not -path "$workspace/README.md" | wc -l | tr -d ' '
}
count_dirs() {
  find "$workspace" -type d -not -path "$workspace/.git" -not -path "$workspace/.git/*" -not -path "$workspace/.ci" -not -path "$workspace/.ci/*" | wc -l | awk '{print $1-1}'
}
line_count() {
  find "$workspace" -type f -name "$1" -not -path "$workspace/.git/*" -not -path "$workspace/.ci/*" -print0 | xargs -0 wc -l 2>/dev/null | awk 'END {print ($1 + 0)}'
}
bytes_in() {
  find "$1" -type f -printf '%s\n' 2>/dev/null | awk '{sum += $1} END {print sum + 0}'
}

go_files=$(find "$workspace" -type f -name '*.go' -not -path "$workspace/.git/*" -not -path "$workspace/.ci/*" | wc -l | tr -d ' ')
gooo_files=$(find "$workspace" -type f -name '*.gooo' -not -path "$workspace/.git/*" -not -path "$workspace/.ci/*" | wc -l | tr -d ' ')
regular_files=$(count_files)
subdirectories=$(count_dirs)
go_lines=$(line_count '*.go')
gooo_lines=$(line_count '*.gooo')
outputs_count=$(find "$ci_dir/runs" "$ci_dir/integration" -type f 2>/dev/null | wc -l | tr -d ' ')
outputs_bytes=$(bytes_in "$ci_dir/runs")
integration_bytes=$(bytes_in "$ci_dir/integration")
outputs_bytes=$((outputs_bytes + integration_bytes))
generated_count=$(find "$ci_dir/runs" -type f \( -name 'generated.go' -o -name 'semantic-ir.json' -o -name 'provenance.json' \) 2>/dev/null | wc -l | tr -d ' ')
generated_bytes=$(find "$ci_dir/runs" -type f \( -name 'generated.go' -o -name 'semantic-ir.json' -o -name 'provenance.json' \) -printf '%s\n' 2>/dev/null | awk '{sum += $1} END {print sum + 0}')

jq -n \
  --slurpfile conformance "$ci_dir/conformance-report.json" \
  --slurpfile compile "$ci_dir/timing/compile.json" \
  --slurpfile build "$ci_dir/timing/build.json" \
  --slurpfile test "$ci_dir/timing/test.json" \
  --slurpfile conformance_time "$ci_dir/timing/conformance.json" \
  --slurpfile integration_time "$ci_dir/timing/integration.json" \
  --argjson go_files "$go_files" --argjson gooo_files "$gooo_files" --argjson go_lines "$go_lines" --argjson gooo_lines "$gooo_lines" \
  --argjson regular_files "$regular_files" --argjson subdirectories "$subdirectories" --argjson outputs_count "$outputs_count" --argjson outputs_bytes "$outputs_bytes" \
  --argjson generated_count "$generated_count" --argjson generated_bytes "$generated_bytes" \
  '{schema:"gooo/ci-evidence/v1",inventory:{go_files:$go_files,gooo_files:$gooo_files,go_physical_lines:$go_lines,gooo_physical_lines:$gooo_lines,regular_files:$regular_files,files_total:$regular_files,subdirectories:$subdirectories,root_readme_inventory_excluded:1},outputs:{count:$outputs_count,bytes:$outputs_bytes},generated_artifacts:{count:$generated_count,bytes:$generated_bytes},wall_ms:{compile:$compile[0].wall_ms,build:$build[0].wall_ms,test:$test[0].wall_ms,conformance:$conformance_time[0].wall_ms,integration:$integration_time[0].wall_ms},resources:{peak_rss_kib:{compile:$compile[0].peak_rss_kib,build:$build[0].peak_rss_kib,test:$test[0].peak_rss_kib,conformance:$conformance_time[0].peak_rss_kib,integration:$integration_time[0].peak_rss_kib},peak_rss_bytes:($compile[0].peak_rss_kib * 1024)},tests:$conformance[0].tests,canonical:$conformance[0].counts,local_validation_commands:0,operational_refutation:"OPERATIONAL_REFUTED",authority_counts:{local:0,remote:1,runtime_write:0,merge_authority:0},improvement:{status:"UNKNOWN",reason:"no external user evidence was supplied; confluence is closed only for the canonical source evidence"}}' > "$output"
