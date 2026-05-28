#!/bin/bash

set -e 
echo "Running language installation scripts..."
SCRIPT_DIR="$(dirname "$0")"
for script in "$SCRIPT_DIR/lang_install"/*.sh; do
    if [ -f "$script" ]; then
        echo "Executing $script..."
        bash "$script"
    fi
done

apt-get clean