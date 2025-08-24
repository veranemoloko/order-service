#!/bin/bash
UID_FILE="send_get_scripts/sample_data/uids.txt"
BASE_URL="http://host.docker.internal:8081/orders"

if [[ ! -f "$UID_FILE" ]]; then
  echo "UID file not found: $UID_FILE"
  exit 1
fi

while IFS= read -r uid; do
  [[ -z "$uid" ]] && continue

  echo "Requesting UID: $uid"

  response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/${uid}")
  
  status=$(echo "$response" | tail -n1)

  body=$(echo "$response" | sed '$d')

  echo "HTTP Status: $status"
  
  if echo "$body" | jq . >/dev/null 2>&1; then
    echo "$body" | jq .
  else
    echo "$body"
  fi

  echo "--------------------------"
  
  sleep 3
done < "$UID_FILE"

echo "All requests completed."
