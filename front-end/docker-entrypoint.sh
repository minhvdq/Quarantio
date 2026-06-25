#!/bin/sh
set -e

TENANT_URL="${TENANT_URL:-}"

cat > /usr/share/nginx/html/config.js << EOF
window.__CONFIG__ = {
  BROKER_URL: "",
  TENANT_URL: "${TENANT_URL}"
};
EOF

exec nginx -g "daemon off;"
