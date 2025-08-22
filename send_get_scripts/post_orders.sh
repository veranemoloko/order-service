#!/bin/bash

docker cp ./send_get_scripts/. sender:/send_get_scripts/

docker exec -d sender /bin/sh /send_get_scripts/send.sh