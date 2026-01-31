# Agromart Deployment Guide (Azure VM)

This guide explains how to deploy the Agromart application to a frugal Azure VM (B-series) using Docker Compose and Caddy for automatic HTTPS.

## Prerequisites

1.  **Azure CLI**: Ensure `az` is installed and logged in (`az login`).
2.  **Domain Name**: You need a domain name (e.g., `demo.agromart.com`) pointing to the VM's IP.
3.  **Resend API Key**: For email verification.

## Step 1: Configuration

1.  Copy the template environment file:
    ```bash
    cp deploy/.env.prod.template deploy/.env.prod
    ```
2.  Edit `deploy/.env.prod` and fill in your details:
    *   **DOMAIN_NAME**: Your domain (e.g., `demo.example.com`).
    *   **RESEND_API_KEY**: Your Resend API key.
    *   **RESEND_FROM_EMAIL**: A verified sender email (e.g., `noreply@demo.example.com`).
    *   **Passwords**: Update database and MinIO passwords.

## Step 2: Provision Azure VM

Run the provisioning script to create a resource group and a B2s VM (approx. $30/mo, burstable).

```bash
./deploy/provision_azure.sh
```

*   **Note**: This script uses `az vm create` with SSH key generation.
*   **Output**: It will display the **Public IP Address** of the new VM.

## Step 3: Configure DNS

Go to your domain registrar (GoDaddy, Namecheap, Cloudflare, etc.) and create an **A Record**:

*   **Host**: `@` or `demo` (depending on your `DOMAIN_NAME`).
*   **Value**: The **Public IP Address** from Step 2.
*   **TTL**: 300 (or lowest possible for fast propagation).

**Wait a few minutes for DNS to propagate.**

## Step 4: Deploy

Run the deployment script with the VM's IP address:

```bash
./deploy/deploy_to_vm.sh <VM_IP_ADDRESS>
```

This script will:
1.  Install Docker on the VM (if missing).
2.  Sync the project code to the VM (excluding `node_modules`).
3.  Build and start the services using Docker Compose.
4.  Prune old images to save disk space.

## Step 5: Verification

1.  Open `https://<YOUR_DOMAIN_NAME>` in your browser.
    *   Caddy will automatically acquire a Let's Encrypt certificate.
    *   It might take a few seconds on the first load.
2.  Test the API: `https://<YOUR_DOMAIN_NAME>/v1/health`.
3.  Test Sign Up: Create an account. Resend should send a verification email.

## Troubleshooting

*   **Logs**: SSH into the VM and check logs:
    ```bash
    ssh azureuser@<VM_IP_ADDRESS>
    cd ~/app/deploy
    sudo docker compose -f docker-compose.prod.yml logs -f
    ```
*   **Database Access**: The database is not exposed publicly. Use a tunnel if you need to access it:
    ```bash
    ssh -L 5432:localhost:5432 azureuser@<VM_IP_ADDRESS>
    ```
