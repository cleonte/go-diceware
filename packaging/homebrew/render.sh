#!/bin/sh
# Usage: render.sh VERSION SHA256SUMS > go-diceware.rb
# Fills packaging/homebrew/go-diceware.rb.in from a release's checksum file.
set -eu
version=$1
sums=$2
dir=$(dirname "$0")
sed -e "s/@VERSION@/$version/g" "$dir/go-diceware.rb.in" > "$dir/.render.tmp"
for p in darwin_arm64 darwin_amd64 linux_arm64 linux_amd64; do
	sum=$(awk -v f="diceware_${version}_$p.tar.gz" '$2 == f {print $1}' "$sums")
	[ -n "$sum" ] || { echo "no checksum for $p" >&2; rm -f "$dir/.render.tmp"; exit 1; }
	sed -i.bak -e "s/@SHA_$p@/$sum/" "$dir/.render.tmp"
done
rm -f "$dir/.render.tmp.bak"
cat "$dir/.render.tmp"
rm -f "$dir/.render.tmp"
