# Quick Start - Azure VM Deployment (Single VM)

Deploy Agromart to a single Azure VM in 5 minutes.

## ⚡ Quick Deploy (TL;DR)

```bash
# 1. Set up Azure credentials (one-time)
az login
az ad sp create-for-rbac --name "github-actions" --role contributor --scopes /subscriptions/$(az account show --query id -o tsv) --sdk-auth
# Copy the ENTIRE JSON output

# 2. Add GitHub Secrets (in GitHub repo → Settings → Secrets)
# AZURE_CREDENTIALS: <paste the JSON from step 1>
# DOCKER_USERNAME: your-dockerhub-username
# DOCKER_PASSWORD: your-dockerhub-token
# RESEND_API_KEY: get from https://resend.com (REQUIRED!)
# JWT_SECRET: $(openssl rand -base64 32)
# CSRF_SECRET: $(openssl rand -base64 32)
# POSTGRES_PASSWORD: $(openssl rand -base64 24)
# MINIO_ROOT_PASSWORD: $(openssl rand -base64 24)
# SUPER_ADMIN_PASSWORD: your-secure-admin-password
# RESEND_FROM_EMAIL: noreply@agromart-demo.com (or onboarding@resend.dev)
# RESEND_FROM_NAME: Agromart Inventory
# FROM_EMAIL: noreply@agromart-demo.com
# POSTGRES_USER: agromart
# POSTGRES_DB: agromart
# MINIO_ROOT_USER: minioadmin
# MINIO_ACCESS_KEY: minioadmin
# SUPER_ADMIN_EMAIL: admin@agromart-demo.com

# 3. Trigger deployment
# Go to GitHub → Actions → "Deploy to Azure VM" → Run workflow
# ✅ Check "Provision new VM" for first deploy

# 4. Wait 5-10 minutes

# 5. Get URL from workflow output
# Example: https://agromart-demo-xxxxx.centralindia.cloudapp.azure.com

# 6. Test email verification
# - Open the URL
# - Sign up with your real email
# - Check inbox for verification email
# - Click link to verify
```

## 🎯 Critical Requirements

### MUST DO (Email Won't Work Without This)
1. **Sign up at https://resend.com**
2. **Create API Key** in Resend Dashboard
3. **Add `RESEND_API_KEY` to GitHub Secrets**

Without a real Resend API key, users cannot verify their email addresses!

## 🏗️ What's Deployed

**Single Azure VM** running all services:

| Service | Port | Purpose |
|---------|------|---------|
| Caddy | 80/443 | Reverse Proxy + HTTPS |
| Frontend | 3000 | Next.js Web App |
| Backend | 8080 | Go API Server |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache/Sessions |
| MinIO | 9000 | File Storage |

**VM Specs:**
- Size: Standard_B2ats_v2 (2 vCPU, 1GB RAM + 4GB Swap)
- OS: Ubuntu 22.04 LTS
- Cost: ~$15-20/month
- Location: Central India (configurable)

## 📋 Pre-Deployment Checklist

- [ ] Docker Hub account created
- [ ] Resend account created with API key
- [ ] Azure CLI installed and logged in
- [ ] All GitHub secrets configured

## 🔗 URLs After Deployment

| Endpoint | URL |
|----------|-----|
| Application | `https://agromart-demo-xxxxx.centralindia.cloudapp.azure.com` |
| API | `https://agromart-demo-xxxxx.centralindia.cloudapp.azure.com/v1` |
| Health | `https://agromart-demo-xxxxx.centralindia.cloudapp.azure.com/v1/health` |

## 🧪 Test Email Verification

1. Open your VM URL
2. Click "Sign Up"
3. Enter your **real** email address
4. Check your email inbox (and spam folder)
5. Look for email from: `noreply@agromart-demo.com`
6. Click "Verify Email Address" button
7. Account should be activated

## 🆘 Common Issues

### "No verification email received"
→ Check GitHub secret `RESEND_API_KEY` is correct and active in Resend dashboard

### "Deployment failed"
→ Check GitHub Actions logs for missing secrets or Azure credential issues

### "VM out of memory"
→ Normal for first deploy. Swap file is created automatically. Retry deployment.

### "Services unhealthy"
→ SSH to VM and check logs: `ssh azureuser@{fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs'`

### "HTTPS not working"
→ Caddy needs a few minutes to provision certificate. Wait 2-3 minutes and refresh.

## 🔧 Management Commands

```bash
# SSH to VM
ssh azureuser@{vm-fqdn}

# View logs
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs -f'

# Restart all services
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml restart'

# Check status
ssh azureuser@{vm-fqdn} 'docker ps'

# View resource usage
ssh azureuser@{vm-fqdn} 'docker stats --no-stream'
```

## 🛑 Stop/Start VM (Save Money)

```bash
# Stop VM (stops billing for compute, keeps storage)
az vm stop -g agromart-demo-rg -n agromart-demo-vm

# Start VM
az vm start -g agromart-demo-rg -n agromart-demo-vm

# Delete VM completely (DESTROYS ALL DATA)
az group delete -n agromart-demo-rg --yes
```

## 📚 Full Documentation

See [AZURE_VM_DEPLOYMENT.md](./AZURE_VM_DEPLOYMENT.md) for complete guide with troubleshooting.

## 💬 Support

- GitHub Issues: Create issue in repository
- Azure Docs: https://docs.microsoft.com/azure/virtual-machines/
- Resend Docs: https://resend.com/docs
