#!/bin/sh

# Optional runtime env vars:
# echo "{\"NODE\": \"$NODE_NAME\", \"POD\": \"$POD_NAME\"}" > /usr/share/nginx/html/k8s_data.json

nginx -g 'daemon off;'