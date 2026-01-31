# Azure Container Apps Deployment Guide

Complete guide for deploying the Agromart Inventory Management System to Azure Container Apps with production email functionality.

---

## 📋 What Was Set Up

### Infrastructure Components Created

1. **GitHub Actions Workflow** (`.github/workflows/deploy-azure.yml`)
   - Automated build and deployment pipeline
   - Builds Docker images for backend (Go) and frontend (Next.js)
   - Deploys 5 services to Azure Container Apps
   - Health checks and deployment verification

2. **Docker Compose** (`deploy/docker-compose.prod.yml`)
   - Production-ready multi-service configuration
   - Resource limits optimized for Azure Container Apps
   - Health checks for all services
   - Removed Caddy (Azure provides built-in SSL/ingress)

3. **Environment Configuration** (`deploy/.env.prod` + `.env.prod.template`)
   - Production environment variables
   - Clear instructions for required values
   - Placeholder for Resend API key (must be updated!)

### Services Deployed

| Service | Image | Port | Purpose |
|---------|-------|------|---------|
| Frontend | Next.js 15 + Bun | 3000 | React web application |
| Backend | Go 1.25 + Echo | 8080 | REST API server |
| PostgreSQL | Postgres 17 Alpine | 5432 | Primary database |
| Redis | Redis 8 Alpine | 6379 | Caching & sessions |
| MinIO | MinIO Latest | 9000/9001 | S3-compatible storage |

---

## 🔐 Required GitHub Secrets

Add these secrets to your GitHub repository (Settings → Secrets and variables → Actions):

### Azure Credentials
| Secret | How to Obtain |
|--------|---------------|
| `AZURE_CREDENTIALS` | Run: `az ad sp create-for-rbac --name "github-actions" --role contributor --scopes /subscriptions/{subscription-id} --sdk-auth` |

### Docker Hub
| Secret | Description |
|--------|-------------|
| `DOCKER_USERNAME` | Your Docker Hub username |
| `DOCKER_PASSWORD` | Your Docker Hub access token |

### Application Secrets (Generate Strong Random Values)
| Secret | Generation Command |
|--------|-------------------|
| `JWT_SECRET` | `openssl rand -base64 32` |
| `CSRF_SECRET` | `openssl rand -base64 32` |
| `POSTGRES_PASSWORD` | `openssl rand -base64 24` |
| `MINIO_ROOT_PASSWORD` | `openssl rand -base64 24` |
| `SUPER_ADMIN_PASSWORD` | Strong password for initial admin |

### Email Configuration (CRITICAL)
| Secret | How to Obtain |
|--------|---------------|
| `RESEND_API_KEY` | Sign up at https://resend.com → API Keys → Create API Key |
| `RESEND_FROM_EMAIL` | Your verified domain (e.g., `noreply@yourdomain.com`) |
| `RESEND_FROM_NAME` | Display name (e.g., `Agromart Inventory`) |
| `FROM_EMAIL` | Same as RESEND_FROM_EMAIL |

### Application Configuration
| Secret | Value |
|--------|-------|
| `NEXT_PUBLIC_API_URL` | `https://agromart-demo.centralindia.azurecontainerapps.io/v1` |
| `POSTGRES_USER` | `agromart` |
| `POSTGRES_DB` | `agromart` |
| `MINIO_ROOT_USER` | `minioadmin` |
| `MINIO_ACCESS_KEY` | `minioadmin` |
| `SUPER_ADMIN_EMAIL` | `admin@agromart-demo.com` |

---

## 🚀 Deployment Steps

### Step 1: Configure Resend for Email (REQUIRED)

Without a real Resend API key, email verification will not work!

1. **Sign up at [Resend](https://resend.com)**
   - Create a free account
   - No credit card required for demo/testing

2. **Create an API Key**
   - Go to Dashboard → API Keys → Create API Key
   - Name it "Agromart Demo"
   - Copy the key (starts with `re_`)

3. **Verify Your Domain** (Optional for demo)
   - For production: Add your domain and verify DNS records
   - For quick demo: Use `onboarding@resend.dev` (limited to your own email)
   - Update `RESEND_FROM_EMAIL` accordingly

4. **Update GitHub Secret**
   - Add `RESEND_API_KEY` to GitHub repository secrets

### Step 2: Set Up Azure Credentials

```bash
# Login to Azure
az login

# Get your subscription ID
az account show --query id -o tsv

# Create service principal for GitHub Actions
az ad sp create-for-rbac \
  --name "agromart-github-actions" \
  --role contributor \
  --scopes /subscriptions/{subscription-id} \
  --sdk-auth

# Copy the entire JSON output and add as AZURE_CREDENTIALS secret
```

### Step 3: Add All GitHub Secrets

Go to your repository → Settings → Secrets and variables → Actions → New repository secret

Add all secrets listed in the table above.

### Step 4: Trigger Deployment

#### Option A: Manual Trigger (Recommended for first deploy)

1. Go to GitHub repository → Actions tab
2. Select "Deploy to Azure Container Apps" workflow
3. Click "Run workflow"
4. Select environment: `demo`
5. Click "Run workflow"

#### Option B: Push to Deploy Branch

```bash
# Create and push to deploy branch
git checkout -b deploy
git push origin deploy
```

### Step 5: Monitor Deployment

1. Watch the GitHub Actions workflow run
2. Check the Azure Portal → Container Apps
3. Wait for all 5 services to show "Running" status
4. Note the assigned URLs from the workflow output

---

## ✅ Verification Checklist

After deployment completes, verify:

### 1. Health Checks
```bash
# Check backend health
curl https://agromart-backend.xxxxxx.centralindia.azurecontainerapps.io/health

# Should return: {"status":"healthy"}
```

### 2. Frontend Loads
- Open frontend URL in browser
- Should see login/signup page

### 3. Email Verification Works (CRITICAL)
1. Go to frontend URL
2. Sign up with a real email address
3. Check your email inbox
4. **You should receive a verification email from Resend**
5. Click the verification link
6. Account should be activated

### 4. Complete User Flow
1. Register new user
2. Verify email
3. Login
4. Create a product
5. Create an order
6. View analytics dashboard

---

## 🔧 Troubleshooting

### Issue: Deployment Fails

**Check GitHub Actions logs:**
- Go to Actions tab → failed workflow → View logs
- Common issues:
  - Missing GitHub secrets
  - Invalid Azure credentials
  - Docker Hub authentication failed

### Issue: Services Unhealthy

**Check Container App logs in Azure Portal:**
1. Go to Azure Portal → Container Apps
2. Select the service
3. Click "Log stream"
4. Look for error messages

**Common causes:**
- Missing environment variables
- Database connection issues
- Incorrect service URLs

### Issue: Emails Not Sending

**Verify Resend Configuration:**
1. Check GitHub secret `RESEND_API_KEY` is set correctly
2. Verify in Resend Dashboard that API key is active
3. Check backend logs for email errors
4. Test with Resend's test domain first: `onboarding@resend.dev`

**Backend logs:**
```bash
# Check backend container logs
az containerapp logs show \
  --name agromart-backend \
  --resource-group agromart-demo-rg \
  --follow
```

### Issue: Frontend Can't Connect to Backend

**Check CORS configuration:**
- Backend should have `FRONTEND_URL` set correctly
- Check browser console for CORS errors
- Verify `NEXT_PUBLIC_API_URL` in frontend build args

---

## 📊 Azure Resource Management

### View Resources
```bash
# List all resources
az resource list --resource-group agromart-demo-rg --output table

# View Container Apps
az containerapp list --resource-group agromart-demo-rg --output table
```

### View Logs
```bash
# Backend logs
az containerapp logs show \
  --name agromart-backend \
  --resource-group agromart-demo-rg \
  --follow

# Frontend logs
az containerapp logs show \
  --name agromart-frontend \
  --resource-group agromart-demo-rg \
  --follow
```

### Delete Resources (Cleanup)
```bash
# Delete entire resource group (removes everything)
az group delete --name agromart-demo-rg --yes
```

---

## 💰 Cost Optimization

Azure Container Apps pricing for this deployment:

| Service | vCPU | Memory | Monthly Cost (approx) |
|---------|------|--------|----------------------|
| Backend | 0.5 | 1GB | ~$15-20 |
| Frontend | 0.25 | 0.5GB | ~$8-12 |
| PostgreSQL | 0.5 | 1GB | ~$15-20 |
| Redis | 0.25 | 0.5GB | ~$8-12 |
| MinIO | 0.25 | 0.5GB | ~$8-12 |
| **Total** | | | **~$55-75/month** |

**Cost saving tips:**
- Use `min-replicas: 0` for non-critical services (scale to zero)
- Delete resources when not in use
- Use Azure free tier if eligible ($200 credit for 30 days)

---

## 🔒 Security Considerations

### Implemented
- ✅ HTTPS enforced (Azure-managed certificates)
- ✅ Secrets stored in GitHub Secrets (not in code)
- ✅ Non-root containers
- ✅ Health checks on all services
- ✅ CSRF protection
- ✅ JWT authentication

### Recommended Additional Steps
- [ ] Enable Azure Key Vault for secrets
- [ ] Configure Azure AD authentication
- [ ] Set up Azure Monitor alerts
- [ ] Enable WAF (Web Application Firewall)
- [ ] Regular security scans

---

## 📝 Next Steps

### For Production Use
1. **Custom Domain**: Configure your own domain in Azure
2. **SSL Certificate**: Azure provides free managed certificates
3. **Monitoring**: Set up Azure Monitor and Application Insights
4. **Backups**: Configure automated database backups
5. **Scaling**: Adjust replica counts based on load

### For Development
1. Use local Docker Compose for development
2. Deploy to Azure only for demos/testing
3. Keep separate environments (dev/staging/prod)

---

## 🆘 Getting Help

### Resources
- [Azure Container Apps Docs](https://docs.microsoft.com/azure/container-apps/)
- [Resend Documentation](https://resend.com/docs)
- [GitHub Actions Docs](https://docs.github.com/actions)

### Common Commands
```bash
# Check deployment status
az containerapp show \
  --name agromart-backend \
  --resource-group agromart-demo-rg \
  --query properties.provisioningState

# Restart a service
az containerapp restart \
  --name agromart-backend \
  --resource-group agromart-demo-rg

# Update environment variable
az containerapp update \
  --name agromart-backend \
  --resource-group agromart-demo-rg \
  --set-env-vars "RESEND_API_KEY=your_new_key"
```

---

## ✅ Definition of Done

- [ ] All 5 services running in Azure Container Apps
- [ ] User can register and receive real verification email
- [ ] User can click verification link and activate account
- [ ] Complete user flow works (login, create product, create order)
- [ ] HTTPS enabled with valid certificate
- [ ] GitHub Actions workflow deploys successfully
- [ ] No secrets exposed in code or logs

---

**Last Updated**: 2026-01-30
**Deployment Version**: 1.0.0
