#!/usr/bin/env bash
set -euo pipefail

output=$1
shift
if [[ "${1:-}" == "--" ]]; then
  shift
fi
mkdir -p "$(dirname "$output")"
rss_file=$(mktemp)
started=$(date +%s%3N)
set +e
/usr/bin/time -f '%M' -o "$rss_file" "$@"
status=$?
set -e
finished=$(date +%s%3N)
peak_rss_kib=$(awk 'NF {print $1; exit}' "$rss_file")
rm -f "$rss_file"
peak_rss_kib=${peak_rss_kib:-0}
jq -n --argjson wall_ms "$((finished - started))" --argjson peak_rss_kib "$peak_rss_kib" --argjson exit_code "$status" \
  '{wall_ms:$wall_ms,peak_rss_kib:$peak_rss_kib,exit_code:$exit_code}' > "$output"
exit "$status"
