set -euo pipefail
README="README.md"
START="<!-- BENCHMARKS:START -->"
END="<!-- BENCHMARKS:END -->"
TMP=$(mktemp)
trap 'rm -f "$TMP"' EXIT

echo "Running benchmarks"
go test -run ^$ -bench=. -benchmem ./... | grep Benchmark > "$TMP"

{
while IFS= read -r line; do
	if [[ "$line" == "$START" ]]; then
		echo "$line"
		echo "| Benchmark | Iterations | ns/op | B/op | allocs/op | "
		echo "| --- | ---: | ---: | ---: | ---: |"
		while read -r name iterations ns_op _ bytes_op _ allocs_op _; do
			echo "$name | $iterations | $ns_op | $bytes_op | $allocs_op |"
		done < "$TMP"
		while IFS= read -r line && [[ "$line" != "$END" ]]; do
			:
		done
		echo "$END"
	else
		echo "$line"
	fi
done < "$README"
} > "${README}.tmp"

mv "${README}.tmp" "$README"
