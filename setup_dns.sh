#!/bin/bash

# Script to add local DNS entries for testing subdomains
# Usage: sudo ./setup_dns.sh

DOMAIN="agromart.local"
TENANTS=("tenant1" "tenant2" "admin" "demo")

echo "Adding entries to /etc/hosts..."

for tenant in "${TENANTS[@]}"; do
    ENTRY="127.0.0.1 $tenant.$DOMAIN"
    if grep -q "$ENTRY" /etc/hosts; then
        echo "Entry '$ENTRY' already exists."
    else
        echo "Adding '$ENTRY'..."
        echo "$ENTRY" | sudo tee -a /etc/hosts
    fi
done

echo "Done. You can now access:"
for tenant in "${TENANTS[@]}"; do
    echo "http://$tenant.$DOMAIN:3000"
done
