#!/bin/sh
VITE_AXIOS_BASE_URL=${VITE_AXIOS_BASE_URL:-"http://localhost:12356/api/v1"}

# replace '__VITE_AXIOS_BASE_URL__' to ${VITE_AXIOS_BASE_URL}
find /usr/share/nginx/html -type f | xargs sed -i "s|__VITE_AXIOS_BASE_URL__|${VITE_AXIOS_BASE_URL}|g"
exec nginx -g 'daemon off;'
