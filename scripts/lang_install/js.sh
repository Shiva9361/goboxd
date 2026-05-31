#!/bin/bash

set -e 

echo "Installing JavaScript (Node.js)..."

apt-get update
apt-get install -y --no-install-recommends nodejs npm

echo "Installed Node version: $(node -v)"
echo "Installed NPM version: $(npm -v)"

echo "Running JavaScript engine test..."

cat << 'EOF' > test.js
const message = "JavaScript Sandbox is functional.";
console.log(message);
EOF

node test.js
rm test.js

echo "JavaScript installation complete!"