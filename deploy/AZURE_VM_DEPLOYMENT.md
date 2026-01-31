# Azure VM Deployment Guide (Single VM with Docker Compose)

Deploy the Agromart Inventory Management System to a single Azure VM with all services running via Docker Compose.

---

## 🏗️ Architecture

All services run on **ONE Azure VM** (Standard_B2ats_v2):

```
┌─────────────────────────────────────────────────────────────┐
│                    Azure VM (Ubuntu 22.04)                 │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Caddy     │  │  Frontend   │  │   Backend   │         │
│  │  (Reverse   │  │  (Next.js)  │  │    (Go)     │         │
│  │   Proxy)    │  │   :3000     │  │   :8080     │         │
│  │   :80/443   │  │             │  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│         │                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  PostgreSQL │  │    Redis    │  │    MinIO    │         │
│  │   :5432     │  │   :6379     │  │   :9000     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**VM Specs:**
- **Size:** Standard_B2ats_v2 (2 vCPU, 1 GB RAM) + 4GB Swap
- **OS:** Ubuntu 22.04 LTS
- **Cost:** ~$15-20/month
- **Location:** Central India (or your choice)

---

## 📋 What Was Set Up

### Infrastructure Components

1. **GitHub Actions Workflow** (`.github/workflows/deploy-azure.yml`)
   - Provisions Azure VM (if needed)
   - Installs Docker and Docker Compose
   - Syncs code and builds on the VM
   - Deploys all 6 services
   - Health checks after deployment

2. **Docker Compose** (`deploy/docker-compose.vm.yml`)
   - All 6 services on single VM
   - Caddy reverse proxy with automatic HTTPS
   - Resource limits for 1GB RAM VM
   - Health checks and restart policies

3. **Caddy Configuration** (`deploy/Caddyfile`)
   - Automatic HTTPS (Let's Encrypt)
   - Routes `/v1/*` to backend
   - Routes everything else to frontend
   - WebSocket support

4. **Environment Files**
   - `deploy/.env.prod` - Production environment
   - `deploy/.env.prod.template` - Template with instructions

---

## 🔐 Required GitHub Secrets

Add these to your repository (Settings → Secrets and variables → Actions):

| Secret | How to Obtain |
|--------|---------------|
| `AZURE_CREDENTIALS` | `az ad sp create-for-rbac --name "github-actions" --role contributor --scopes /subscriptions/{subscription-id} --sdk-auth` |
| `DOCKER_USERNAME` | Your Docker Hub username (for image pulls if needed) |
| `DOCKER_PASSWORD` | Your Docker Hub access token |

### Application Secrets

| Secret | Generation Command |
|--------|-------------------|
| `JWT_SECRET` | `openssl rand -base64 32` |
| `CSRF_SECRET` | `openssl rand -base64 32` |
| `POSTGRES_PASSWORD` | `openssl rand -base64 24` |
| `MINIO_ROOT_PASSWORD` | `openssl rand -base64 24` |
| `SUPER_ADMIN_PASSWORD` | Choose a strong password |

### Email Configuration (⚠️ CRITICAL)

| Secret | How to Obtain |
|--------|---------------|
| `RESEND_API_KEY` | Sign up at https://resend.com → API Keys |
| `RESEND_FROM_EMAIL` | Your verified domain or `onboarding@resend.dev` |
| `RESEND_FROM_NAME` | Display name (e.g., "Agromart Inventory") |
| `FROM_EMAIL` | Same as RESEND_FROM_EMAIL |

### Application Configuration

| Secret | Value |
|--------|-------|
| `NEXT_PUBLIC_API_URL` | Will be auto-set to `https://{vm-fqdn}/v1` |
| `POSTGRES_USER` | `agromart` |
| `POSTGRES_DB` | `agromart` |
| `MINIO_ROOT_USER` | `minioadmin` |
| `MINIO_ACCESS_KEY` | `minioadmin` |
| `SUPER_ADMIN_EMAIL` | `admin@agromart-demo.com` |

---

## 🚀 Deployment Steps

### Step 1: Configure Resend for Email (REQUIRED)

Without a real Resend API key, email verification **will not work**:

1. **Sign up at https://resend.com**
   - Free account, no credit card needed for demo

2. **Create API Key**
   - Dashboard → API Keys → Create API Key
   - Copy the key (starts with `re_`)

3. **Update GitHub Secrets**
   - Add `RESEND_API_KEY` to repository secrets

### Step 2: Set Up Azure Credentials

```bash
# Login to Azure
az login

# Get subscription ID
SUBSCRIPTION_ID=$(az account show --query id -o tsv)

# Create service principal
az ad sp create-for-rbac \
  --name "agromart-github-actions" \
  --role contributor \
  --scopes /subscriptions/$SUBSCRIPTION_ID \
  --sdk-auth

# Copy the ENTIRE JSON output and add as AZURE_CREDENTIALS secret
```

### Step 3: Add All GitHub Secrets

Go to repository → Settings → Secrets and variables → Actions → New repository secret

Add all 15+ secrets from the tables above.

### Step 4: Trigger Deployment

**Option A: Provision New VM + Deploy** (First time)

1. Go to GitHub → Actions → "Deploy to Azure VM"
2. Click "Run workflow"
3. Check "Provision new VM (destroys existing)"
4. Click "Run workflow"

**Option B: Deploy to Existing VM** (Updates)

1. Go to GitHub → Actions → "Deploy to Azure VM"
2. Click "Run workflow"
3. **UNCHECK** "Provision new VM"
4. Click "Run workflow"

**Option C: Push to Deploy Branch**

```bash
git checkout -b deploy-vm
git push origin deploy-vm
```

### Step 5: Wait & Get URL

- Deployment takes **5-10 minutes** (provisioning VM) or **2-3 minutes** (update)
- Watch GitHub Actions for the VM FQDN (e.g., `agromart-demo-xxxxx.centralindia.cloudapp.azure.com`)
- Your app will be at: `https://{vm-fqdn}`

---

## ✅ Verification Checklist

After deployment:

### 1. Access Application
```bash
# The workflow will output the URL
# Example: https://agromart-demo-xxxxx.centralindia.cloudapp.azure.com
```

### 2. Test Health Endpoints
```bash
# Frontend
curl https://{vm-fqdn}

# Backend
curl https://{vm-fqdn}/v1/health
# Should return: {"status":"healthy"}
```

### 3. Test Email Verification (CRITICAL)
1. Open the application URL
2. Click "Sign Up"
3. Enter your **real email address**
4. Check your email inbox
5. **You should receive a verification email from Resend**
6. Click the verification link
7. Account should be activated

### 4. Complete User Flow
1. Register new user
2. Verify email (check inbox)
3. Login
4. Create a product
5. Create an order
6. View analytics dashboard

---

## 🔧 Management Commands

### SSH into VM
```bash
ssh azureuser@{vm-fqdn}
```

### View Logs
```bash
# All services
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs -f'

# Specific service
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs -f backend'
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs -f frontend'
```

### Restart Services
```bash
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml restart'
```

### Update (Pull Latest)
```bash
# Re-trigger GitHub Actions workflow (recommended)
# Or manually:
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml pull && docker compose -f docker-compose.vm.yml up -d'
```

### Check Resource Usage
```bash
ssh azureuser@{vm-fqdn} 'docker stats --no-stream'
ssh azureuser@{vm-fqdn} 'free -h'
ssh azureuser@{vm-fqdn} 'df -h'
```

---

## 🔧 Troubleshooting

### Issue: VM Out of Memory

**Symptoms:** Services crash, OOM errors in logs

**Solution:**
```bash
# SSH to VM and check
ssh azureuser@{vm-fqdn} 'free -h'

# If swap is not active:
ssh azureuser@{vm-fqdn} 'sudo swapon /swapfile'

# Reduce memory limits in docker-compose.vm.yml
# Edit: deploy/docker-compose.vm.yml
# Lower memory limits for non-critical services
```

### Issue: Deployment Fails

**Check GitHub Actions logs** for:
- Missing secrets
- Azure credentials issues
- SSH connection failures

**Common fixes:**
```bash
# Regenerate Azure credentials
az ad sp create-for-rbac --name "github-actions" --role contributor --scopes /subscriptions/{id} --sdk-auth
```

### Issue: Services Unhealthy

```bash
# Check individual service status
ssh azureuser@{vm-fqdn} 'docker ps'

# Check logs
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs backend'

# Common issues:
# - Database not ready: Wait and restart
# - Missing env vars: Check .env file on VM
ssh azureuser@{vm-fqdn} 'cat ~/app/deploy/.env'
```

### Issue: Emails Not Sending

**Verify Resend Configuration:**
1. Check GitHub secret `RESEND_API_KEY` is set
2. Verify in Resend Dashboard that API key is active
3. Check backend logs for email errors:
```bash
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs backend | grep -i email'
```

### Issue: HTTPS Not Working

Caddy automatically provisions HTTPS, but it may take a few minutes.

```bash
# Check Caddy logs
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml logs caddy'

# Restart Caddy if needed
ssh azureuser@{vm-fqdn} 'cd ~/app/deploy && docker compose -f docker-compose.vm.yml restart caddy'
```

---

## 💰 Cost Optimization

**Monthly Costs (~$15-20):**
- VM (Standard_B2ats_v2): ~$15/month
- Bandwidth: Minimal for demo
- Storage: ~$1-2/month

**Cost Saving Tips:**
- Stop VM when not in use: `az vm stop -g agromart-demo-rg -n agromart-demo-vm`
- Use Azure free tier ($200 credit for 30 days)
- Delete VM entirely when done: `az vm delete -g agromart-demo-rg -n agromart-demo-vm --yes`

---

## 🔒 Security

### Implemented
- ✅ HTTPS via Caddy (auto Let's Encrypt)
- ✅ Secrets stored in GitHub (not in code)
- ✅ Non-root Docker containers
- ✅ SSH key authentication
- ✅ CSRF protection
- ✅ JWT authentication

### Additional Recommendations
- [ ] Restrict SSH access by IP
- [ ] Enable Azure NSG rules
- [ ] Regular security updates: `ssh azureuser@{vm-fqdn} 'sudo apt update && sudo apt upgrade -y'`
- [ ] Set up Azure Backup

---

## 📝 Files Reference

| File | Purpose |
|------|---------|
| `.github/workflows/deploy-azure.yml` | GitHub Actions deployment workflow |
| `deploy/docker-compose.vm.yml` | Docker Compose for single VM |
| `deploy/Caddyfile` | Caddy reverse proxy config |
| `deploy/.env.prod` | Production environment variables |
| `deploy/.env.prod.template` | Template with instructions |

---

## 🆘 Getting Help

### View VM Status
```bash
az vm show -g agromart-demo-rg -n agromart-demo-vm -d --query '{IP:publicIps, FQDN:fqdns, State:powerState}'
```

### Restart VM
```bash
az vm restart -g agromart-demo-rg -n agromart-demo-vm
```

### Delete Everything
```bash
# WARNING: This deletes the VM and all data
az group delete -n agromart-demo-rg --yes
```

---

## ✅ Definition of Done

- [ ] VM provisioned and running
- [ ] All 6 services healthy (caddy, frontend, backend, postgres, redis, minio)
- [ ] HTTPS working with valid certificate
- [ ] User can register and receive **real** verification email
- [ ] User can click verification link and activate account
- [ ] Complete user flow works (login, create product, create order)
- [ ] No secrets exposed in code or logs

---

**Last Updated:** 2026-01-30
**Deployment Version:** 2.0.0 (Single VM)
