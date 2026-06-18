#!/bin/sh
set -e

# In production, API calls are proxied by nginx — no absolute URLs needed.
# config.js is a no-op but kept so the <script> tag in index.html doesn't 404.
cat > /usr/share/nginx/html/config.js << 'EOF'
window.__CONFIG__ = {};
EOF

exec nginx -g "daemon off;"
