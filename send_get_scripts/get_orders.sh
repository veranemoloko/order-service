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

  status=$(curl -s -o >(cat) -w "%{http_code}" "${BASE_URL}/${uid}")
  
  echo "HTTP Status: $status"
  echo "--------------------------"
  
  sleep 3
done < "$UID_FILE"

echo "All requests completed."
