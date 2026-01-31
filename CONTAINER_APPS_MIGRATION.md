# Azure Container Apps Migration - Complete! ✅

Successfully migrated from Azure VM deployment to Azure Container Apps.

## What Changed

### ✅ New Files
- `.github/workflows/deploy-container-apps.yml` - New Container Apps deployment workflow
- `.github/workflows/deploy-azure-vm.yml.bak` - Archived VM workflow (for reference)

### ✅ Updated Files
- `.github/workflows/ci.yml` - Now deploys to Container Apps on push to main
- `deploy/AZURE_DEPLOYMENT.md` - Updated to reference correct workflow file

## How It Works

When you push to the `main` branch:

1. **CI Pipeline runs** (tests, lints, builds)
2. **Docker images are built** and pushed to Docker Hub
3. **Container Apps deployment triggers** automatically
4. **5 services are deployed**:
   - PostgreSQL 17 (database)
   - Redis 8 (cache)
   - MinIO (object storage)
   - Backend API (Go)
   - Frontend (Next.js)
5. **Health checks verify** all services are running
6. **URLs are outputted** in GitHub Actions summary

## Deployment Targets

### Infrastructure Services (Internal)
- **PostgreSQL**: Internal ingress, 0.5 vCPU, 1GB RAM
- **Redis**: Internal ingress, 0.25 vCPU, 0.5GB RAM
- **MinIO**: Internal ingress, 0.25 vCPU, 0.5GB RAM

### Application Services (External)
- **Backend**: External ingress (HTTPS), 0.5 vCPU, 1GB RAM
  - Auto-scales: 1-3 replicas
  - URL: `https://agromart-backend.centralindia.azurecontainerapps.io`
- **Frontend**: External ingress (HTTPS), 0.25 vCPU, 0.5GB RAM
  - Auto-scales: 1-3 replicas
  - URL: `https://agromart-frontend.centralindia.azurecontainerapps.io`

## Required GitHub Secrets

All existing secrets are still used:

### Azure & Docker
- `AZURE_CREDENTIALS` - Azure service principal
- `DOCKER_USERNAME` - Docker Hub username
- `DOCKER_PASSWORD` - Docker Hub password

### Security
- `JWT_SECRET` - JWT signing key
- `CSRF_SECRET` - CSRF token secret

### Database
- `POSTGRES_USER` - Database username
- `POSTGRES_PASSWORD` - Database password
- `POSTGRES_DB` - Database name

### Storage
- `MINIO_ROOT_USER` - MinIO access key
- `MINIO_ROOT_PASSWORD` - MinIO secret key

### Email (Resend)
- `RESEND_API_KEY` - Resend API key
- `RESEND_FROM_EMAIL` - Sender email address
- `RESEND_FROM_NAME` - Sender display name
- `FROM_EMAIL` - Reply-to email

### Admin
- `SUPER_ADMIN_EMAIL` - Super admin email
- `SUPER_ADMIN_PASSWORD` - Super admin password

## Manual Deployment

To deploy manually without pushing to main:

```bash
# Go to GitHub Actions
# Click "Deploy to Azure Container Apps"
# Click "Run workflow"
# Select environment: production
# Click "Run workflow"
```

## Monitoring & Management

### View Container Apps
```bash
az containerapp list --resource-group agromart-rg --output table
```

### View Logs
```bash
# Backend logs
az containerapp logs show \
  --name agromart-backend \
  --resource-group agromart-rg \
  --follow

# Frontend logs
az containerapp logs show \
  --name agromart-frontend \
  --resource-group agromart-rg \
  --follow
```

### Restart Services
```bash
# Restart backend
az containerapp restart \
  --name agromart-backend \
  --resource-group agromart-rg

# Restart frontend
az containerapp restart \
  --name agromart-frontend \
  --resource-group agromart-rg
```

### Update Environment Variables
```bash
az containerapp update \
  --name agromart-backend \
  --resource-group agromart-rg \
  --set-env-vars "NEW_VAR=value"
```

### Delete All Resources (Cleanup)
```bash
az group delete --name agromart-rg --yes
```

## Benefits Over VM Deployment

### ✅ No SSH Required
- No SSH keys to manage
- No server access needed
- No manual server setup

### ✅ Managed Infrastructure
- Azure handles scaling
- Automatic HTTPS/TLS certificates
- Built-in load balancing
- Health monitoring

### ✅ Simpler Operations
- No Docker installation on VM
- No manual Docker Compose
- No Caddy configuration
- Direct container deployment

### ✅ Better Reliability
- Container restarts on failure
- Health probes monitor services
- Multiple replicas for availability
- No VM boot issues

### ✅ Better Scaling
- Auto-scale based on load (1-3 replicas)
- Scale to zero for cost savings (optional)
- Independent scaling per service

## Cost Comparison

### VM Deployment (Previous)
- **VM**: Standard_B2ats_v2 (~$15-20/month)
- **Total**: ~$15-20/month
- **Issue**: SSH problems, manual management

### Container Apps (Current)
- **Backend**: 0.5 vCPU, 1GB RAM (~$15-20/month)
- **Frontend**: 0.25 vCPU, 0.5GB RAM (~$8-12/month)
- **PostgreSQL**: 0.5 vCPU, 1GB RAM (~$15-20/month)
- **Redis**: 0.25 vCPU, 0.5GB RAM (~$8-12/month)
- **MinIO**: 0.25 vCPU, 0.5GB RAM (~$8-12/month)
- **Total**: ~$55-75/month
- **Benefit**: Much more reliable, no SSH issues, managed infrastructure

## Testing the Deployment

After pushing to main:

1. **Check GitHub Actions** - View deployment progress
2. **Get URLs** - From Actions summary or Azure Portal
3. **Test Frontend** - Visit frontend URL
4. **Test Backend** - Check `/health` endpoint
5. **Test Registration** - Sign up with real email
6. **Verify Email** - Check Resend delivery
7. **Complete Flow** - Login, create products, orders

## Rollback (If Needed)

If Container Apps deployment has issues:

```bash
# Restore VM workflow
mv .github/workflows/deploy-azure-vm.yml.bak .github/workflows/deploy-azure.yml

# Update ci.yml to reference deploy-azure.yml
# Push changes

# Or manually deploy VM workflow
```

## Next Steps

1. ✅ Push to main to trigger first Container Apps deployment
2. ✅ Monitor deployment in GitHub Actions
3. ✅ Test complete application flow
4. ✅ Verify email delivery works
5. ✅ Check all features function correctly
6. ✅ Consider adding custom domain (optional)

## Documentation

- Full deployment guide: `deploy/AZURE_DEPLOYMENT.md`
- VM deployment (archived): `.github/workflows/deploy-azure-vm.yml.bak`
- Container Apps workflow: `.github/workflows/deploy-container-apps.yml`

---

**Migration Date**: 2026-01-31
**Status**: ✅ Complete and ready for deployment
