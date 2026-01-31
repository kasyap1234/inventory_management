#!/bin/bash
set -e

IP_ADDRESS=$1
USER="azureuser"

if [ -z "$IP_ADDRESS" ]; then
  echo "Usage: $0 <VM_IP_ADDRESS>"
  exit 1
fi

if [ ! -f "deploy/.env.prod" ]; then
  echo "Error: deploy/.env.prod not found."
  echo "Please copy deploy/.env.prod.template to deploy/.env.prod and fill in your values."
  exit 1
fi

echo "Deploying to $USER@$IP_ADDRESS..."

# 1. Setup Swap (Critical for small VMs)
echo "Setting up Swap..."
ssh $USER@$IP_ADDRESS <<EOF
  if [ ! -f /swapfile ]; then
    echo "Creating 4GB swapfile..."
    sudo fallocate -l 4G /swapfile
    sudo chmod 600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
    echo "Swap created."
  else
    echo "Swap already exists."
  fi
  free -h
EOF

# 2. Install Docker on VM if missing
echo "Setting up Docker on remote VM..."
ssh -o StrictHostKeyChecking=no $USER@$IP_ADDRESS <<EOF
  if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker $USER
    echo "Docker installed."
  else
    echo "Docker already installed."
  fi
EOF

# Sync files
# We exclude heavy/unnecessary folders
echo "Syncing project files..."
rsync -avz --exclude 'node_modules' --exclude '.git' --exclude '.gemini' --exclude 'tmp' --exclude '.next' \
  ./ $USER@$IP_ADDRESS:~/app/

# Debug: Check remote Dockerfile
ssh $USER@$IP_ADDRESS "cat ~/app/frontend/Dockerfile | grep groupadd"

# 3. Start Application
echo "Starting application..."
ssh $USER@$IP_ADDRESS <<EOF
  cd ~/app/deploy
  
  # Rename .env.prod to .env for Docker Compose variable substitution
  cp .env.prod .env
  
  # Start services
  # We use --build to ensure latest code is running
  sudo docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
  
  # Prune unused images to save space on frugal VM
  sudo docker image prune -f
EOF

echo "--------------------------------------------------"
echo "Deployment Complete!"
echo "Check logs with: ssh $USER@$IP_ADDRESS 'cd ~/app/deploy && sudo docker compose -f docker-compose.prod.yml logs -f'"
echo "--------------------------------------------------"
