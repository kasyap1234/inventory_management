#!/bin/bash
set -e

# Configuration
RG_NAME="rg-agromart-demo-in"
LOCATION="centralindia"
VM_NAME="vm-agromart-demo"
ADMIN_USERNAME="azureuser"
VM_IMAGE="Ubuntu2204"
VM_SIZE="Standard_B2ats_v2" # Frugal choice (AMD)

# Generate a unique DNS label to avoid conflicts
RANDOM_ID=$(openssl rand -hex 4)
DNS_LABEL="agromart-demo-$RANDOM_ID"

echo "Creating Resource Group: $RG_NAME in $LOCATION..."
az group create --name $RG_NAME --location $LOCATION

echo "Creating VM: $VM_NAME ($VM_SIZE) with DNS: $DNS_LABEL..."
# Using generate-ssh-keys to automatically create/use SSH keys
az vm create \
  --resource-group $RG_NAME \
  --name $VM_NAME \
  --image $VM_IMAGE \
  --admin-username $ADMIN_USERNAME \
  --generate-ssh-keys \
  --size $VM_SIZE \
  --public-ip-sku Standard \
  --public-ip-address-dns-name $DNS_LABEL

echo "Opening ports 80 (HTTP) and 443 (HTTPS)..."
az vm open-port --port 80 --resource-group $RG_NAME --name $VM_NAME --priority 1001
az vm open-port --port 443 --resource-group $RG_NAME --name $VM_NAME --priority 1002

echo "Getting Connection Details..."
IP_ADDRESS=$(az vm show --resource-group $RG_NAME --name $VM_NAME -d --query publicIps -o tsv)
FQDN=$(az vm show --resource-group $RG_NAME --name $VM_NAME -d --query fqdns -o tsv)

echo "--------------------------------------------------"
echo "Provisioning Complete!"
echo "VM IP Address: $IP_ADDRESS"
echo "DNS Name (FQDN): $FQDN"
echo "SSH Command: ssh $ADMIN_USERNAME@$FQDN"
echo "--------------------------------------------------"

# Save FQDN to a temp file for the next step to pick up
echo $FQDN > deploy/current_fqdn.txt
echo $IP_ADDRESS > deploy/current_ip.txt

