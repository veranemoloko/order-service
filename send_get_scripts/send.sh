#!/bin/sh

cd /send_get_scripts || exit 1

for file in sample_data/*.json; do
  [ -f "$file" ] || continue
  jq -c . "$file" | kcat -b kafka:29092 -t orders -P
  sleep 10
done