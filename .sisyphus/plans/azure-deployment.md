# Azure Deployment Plan: Agromart Inventory Management System

## TL;DR

> Deploy Agromart Inventory Management System (Go backend + Next.js frontend) to **Azure Container Apps** with production Resend email functionality. All services (PostgreSQL, Redis, MinIO, Backend, Frontend) run as containers with automated GitHub Actions deployment pipeline.
> 
> **Deliverables:**
> - Production-ready Docker Compose setup for Azure
> - GitHub Actions workflow for automated Azure deployment
> - Environment configuration with real Resend API key
> - Domain verification setup for email deliverability
> - Automated SSL/TLS configuration
> 
> **Estimated Effort:** Large (6-8 hours)
> **Parallel Execution:** YES - 3 waves
> **Critical Path:** Phase 1 → Phase 2 → Phase 3 (sequential)

---

## Context

### Original Request
Deploy Agromart Inventory Management System to Azure with production email functionality using Resend for email verification.

### Current State Analysis
**Backend:**
- Go 1.25+ with Echo framework
- Resend email integration (`github.com/resend/resend-go/v2` v2.27.0)
- Multi-tenant architecture with tenant-scoped queries
- Email verification flow with 24hr Redis TTL tokens
- SMTP fallback configured but Resend is primary

**Frontend:**
- Next.js 15 with App Router
- Bun-based build system
- TanStack Query for data fetching
- Standalone output for containerization

**Infrastructure (Current):**
- Local dev: `docker-compose.yml` with PostgreSQL 17, Redis 8, MinIO, Mailpit
- Production template: `deploy/docker-compose.prod.yml` with Caddy reverse proxy
- CI/CD: `.github/workflows/ci.yml` builds and pushes to Docker Hub
- Missing: Azure deployment workflow

**Email Implementation:**
- `internal/services/email_service.go`: Verification email with retry logic (3 attempts, exponential backoff)
- `internal/services/notification_service.go`: Multi-provider email (Resend primary, SMTP fallback)
- Current `.env.example`: Placeholder API key `re_Dr6PobGf_9thMpPV5wRFHgNBG3Y71Ar7b`
- From address: `onboarding@resend.dev` (needs domain verification)

### Key Assumptions (Documented for Review)
1. **Azure Service:** Container Apps (not AKS or VMs) - simplest for Docker Compose workloads
2. **Deployment Strategy:** All services as containers (not managed Azure services) for demo simplicity
3. **Resend Domain:** User will verify a domain (e.g., `agromart-demo.com`) or use Resend's testing domain initially
4. **Region:** East US (changeable via parameter)
5. **Secrets Management:** GitHub Secrets + Azure Container Apps secrets
6. **SSL:** Azure-managed certificates (auto-provisioned)
7. **Custom Domain:** Optional - can use Azure-provided URLs (`*.azurecontainerapps.io`)
8. **Database:** Containerized PostgreSQL (with persistent volume) - NOT production-grade but acceptable for demo
9. **Storage:** Containerized MinIO - NOT production-grade but acceptable for demo
10. **Trigger:** Manual `workflow_dispatch` with optional auto-deploy on main branch

### Guardrails (What NOT to Build)
- ❌ Azure Kubernetes Service (AKS) - too complex for demo
- ❌ Azure managed PostgreSQL/Redis/Storage - overkill for demo, higher cost
- ❌ Multi-environment (staging/prod) - single demo environment only
- ❌ Blue/green deployment - out of scope for demo
- ❌ CDN/Edge optimization - not needed for demo
- ❌ Monitoring/Alerting infrastructure (Grafana, Prometheus) - out of scope
- ❌ Database migration automation beyond basic job
- ❌ Backup/restore automation - out of scope for demo

---

## Work Objectives

### Core Objective
Enable production email sending via Resend and deploy the complete Agromart stack to Azure Container Apps with automated CI/CD.

### Concrete Deliverables
1. Updated `.github/workflows/deploy-azure.yml` - GitHub Actions workflow
2. Updated `deploy/docker-compose.prod.yml` - Azure-compatible production compose
3. Updated `deploy/.env.prod` - Production environment template with Resend config
4. Azure Container Apps environment configuration
5. Resend domain verification setup documentation
6. GitHub Secrets configuration guide
7. Deployment verification procedures

### Definition of Done
- [ ] Azure Container Apps environment created and running
- [ ] All 5 services healthy (frontend, backend, postgres, redis, minio)
- [ ] User can register and receive real verification email
- [ ] User can click verification link and activate account
- [ ] GitHub Actions workflow successfully deploys on trigger
- [ ] HTTPS accessible with valid SSL certificate
- [ ] Environment variables properly secured (no secrets in code)

### Must Have
- Real Resend API key configured (not placeholder)
- Verified sending domain in Resend
- Azure Container Apps deployment
- Automated deployment via GitHub Actions
- SSL/TLS termination
- Health checks for all services
- Secrets management (no hardcoded credentials)

### Must NOT Have (Explicit Exclusions)
- Azure managed database services (Cosmos DB, Azure SQL)
- Kubernetes orchestration
- Multi-region deployment
- Advanced monitoring/observability stack
- Load balancing beyond Azure's built-in
- Custom domain (optional - can use Azure URLs)

---

## Verification Strategy

### Test Decision
- **Infrastructure exists:** YES (Docker Compose)
- **User wants tests:** Manual verification only (no automated tests for deployment)
- **Framework:** N/A - deployment verification
- **QA approach:** Manual verification procedures

### Verification Approach
Each task includes executable verification commands that can be run via bash/shell:
- **Azure CLI commands** for resource verification
- **Docker commands** for container health checks  
- **curl/httpie** for API endpoint verification
- **Manual steps** for email verification (user intervention required)

### Evidence Requirements
- Azure resource list showing all created resources
- Container logs showing successful startup
- Email delivery confirmation (screenshot or Resend dashboard)
- HTTP response codes and headers from endpoints
- SSL certificate validation output

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - No Dependencies):
├── Task 1: Resend Setup (Domain + API Key)
├── Task 2: Azure Infrastructure Setup (Resource Group, Container Apps Environment)
└── Task 3: GitHub Secrets Configuration

Wave 2 (Configuration - After Wave 1):
├── Task 4: Update Docker Compose for Azure
├── Task 5: Create GitHub Actions Workflow
├── Task 6: Update Backend Email Configuration
└── Task 7: Production Environment File Updates

Wave 3 (Deployment - After Wave 2):
├── Task 8: Deploy to Azure
└── Task 9: Verification & Testing

Critical Path: Task 1 → Task 2 → Task 4 → Task 8
Parallel Speedup: ~35% faster than sequential
```

### Dependency Matrix

| Task | Depends On | Blocks | Can Parallelize With |
|------|------------|--------|---------------------|
| 1 (Resend Setup) | None | 6, 7 | 2, 3 |
| 2 (Azure Infra) | None | 4, 8 | 1, 3 |
| 3 (GitHub Secrets) | None | 5, 8 | 1, 2 |
| 4 (Docker Compose) | 2 | 8 | 5, 6, 7 |
| 5 (GitHub Workflow) | 3 | 8 | 4, 6, 7 |
| 6 (Email Config) | 1 | 8 | 4, 5, 7 |
| 7 (Env Files) | 1 | 8 | 4, 5, 6 |
| 8 (Deploy) | 3, 4, 5, 6, 7 | 9 | None |
| 9 (Verification) | 8 | None | None |

### Agent Dispatch Summary

| Wave | Tasks | Recommended Agents | Parallel |
|------|-------|-------------------|----------|
| 1 | 1, 2, 3 | `quick` + `git-master` | YES |
| 2 | 4, 5, 6, 7 | `unspecified-high` + `git-master` | YES |
| 3 | 8, 9 | `unspecified-high` + `dev-browser` | NO (sequential) |

---

## TODOs

### Wave 1: Foundation Setup

- [ ] **1. Resend Domain and API Key Setup**

  **What to do:**
  - Sign up for Resend account at https://resend.com
  - Create a sending domain (or use existing domain)
  - Verify domain via DNS TXT record
  - Generate production API key
  - Document the API key and verified from-email address

  **Must NOT do:**
  - Use the placeholder API key in production
  - Skip domain verification (emails will go to spam)
  - Use onboarding@resend.dev in production (sandbox only)

  **Recommended Agent Profile:**
  - **Category:** `quick`
  - **Skills:** None (external service setup)
  - **Justification:** Simple configuration task requiring web UI interaction

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 1
  - **Blocks:** Task 6 (Email Config), Task 7 (Env Files)
  - **Blocked By:** None

  **References:**
  - Resend Dashboard: https://resend.com/dashboard
  - Domain Verification Docs: https://resend.com/docs/dashboard/domains/introduction
  - API Keys: https://resend.com/dashboard/api-keys

  **Acceptance Criteria:**
  - [ ] Resend account created
  - [ ] Domain added and DNS records configured
  - [ ] Domain shows "Verified" status in Resend dashboard
  - [ ] Production API key generated (format: `re_xxxxxxxx`)
  - [ ] From email address confirmed (e.g., `noreply@yourdomain.com`)

  **Evidence:**
  - Screenshot of Resend dashboard showing verified domain
  - API key stored securely (will be added to GitHub secrets)

  **Commit:** NO (external configuration only)

---

- [ ] **2. Azure Infrastructure Setup**

  **What to do:**
  - Create Azure Resource Group
  - Create Azure Container Apps Environment
  - Configure Log Analytics workspace for monitoring
  - Set up Azure Container Registry (or use Docker Hub)
  - Document resource names and connection strings

  **Must NOT do:**
  - Create Azure SQL Database (use containerized PostgreSQL)
  - Create Azure Cache for Redis (use containerized Redis)
  - Use Kubernetes (keep it simple with Container Apps)

  **Recommended Agent Profile:**
  - **Category:** `unspecified-high`
  - **Skills:** None (Azure CLI work)
  - **Justification:** Infrastructure setup requires Azure CLI knowledge

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 1
  - **Blocks:** Task 4 (Docker Compose), Task 8 (Deploy)
  - **Blocked By:** None

  **References:**
  - Azure Container Apps Docs: https://docs.microsoft.com/azure/container-apps/
  - Azure CLI Reference: https://docs.microsoft.com/cli/azure/

  **Acceptance Criteria:**
  - [ ] Azure CLI authenticated (`az login`)
  - [ ] Resource group created: `az group create --name agromart-rg --location eastus`
  - [ ] Container Apps environment created:
    ```bash
    az containerapp env create \
      --name agromart-env \
      --resource-group agromart-rg \
      --location eastus
    ```
  - [ ] Log Analytics workspace created and linked
  - [ ] Resource IDs documented for GitHub Actions

  **Evidence:**
  ```bash
  # Verify resources exist
  az group show --name agromart-rg
  az containerapp env show --name agromart-env --resource-group agromart-rg
  ```

  **Commit:** NO (infrastructure created externally)

---

- [ ] **3. GitHub Secrets Configuration**

  **What to do:**
  - Add `RESEND_API_KEY` to GitHub repository secrets
  - Add `AZURE_CREDENTIALS` (service principal JSON)
  - Add `DOCKER_USERNAME` and `DOCKER_PASSWORD` (if not exists)
  - Add `JWT_SECRET` (generate random 256-bit string)
  - Add `CSRF_SECRET` (generate random string)
  - Add `DATABASE_URL` components (or full connection string)
  - Add `REDIS_URL`
  - Add `MINIO credentials`

  **Must NOT do:**
  - Commit secrets to the repository
  - Use weak/random secrets for JWT/CSRF
  - Share service principal credentials

  **Recommended Agent Profile:**
  - **Category:** `quick`
  - **Skills:** `git-master`
  - **Justification:** Repository configuration task

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 1
  - **Blocks:** Task 5 (GitHub Workflow), Task 8 (Deploy)
  - **Blocked By:** Task 1 (need Resend API key), Task 2 (need Azure credentials)

  **References:**
  - GitHub Secrets Docs: https://docs.github.com/actions/security-guides/encrypted-secrets
  - Azure Service Principal: https://docs.microsoft.com/cli/azure/ad/sp

  **Acceptance Criteria:**
  - [ ] GitHub repository secrets page accessible
  - [ ] Secrets created:
    - `RESEND_API_KEY`: Real API key from Task 1
    - `AZURE_CREDENTIALS`: Service principal JSON
    - `DOCKER_USERNAME`: Docker Hub username
    - `DOCKER_PASSWORD`: Docker Hub token
    - `JWT_SECRET`: Generated secure random string (32+ chars)
    - `CSRF_SECRET`: Generated secure random string
    - `POSTGRES_PASSWORD`: Generated secure password
    - `MINIO_ROOT_PASSWORD`: Generated secure password
    - `SUPER_ADMIN_PASSWORD`: Generated secure password

  **Evidence:**
  - Screenshot of GitHub Secrets page (secrets masked)
  - JSON output of Azure service principal (redacted):
    ```bash
    az ad sp create-for-rbac \
      --name "agromart-github-actions" \
      --role contributor \
      --scopes /subscriptions/{subscription-id}/resourceGroups/agromart-rg \
      --sdk-auth
    ```

  **Commit:** NO (secrets configured in GitHub UI)

---

### Wave 2: Configuration Updates

- [ ] **4. Update Docker Compose for Azure**

  **What to do:**
  - Update `deploy/docker-compose.prod.yml` for Azure Container Apps compatibility
  - Replace Caddy with Azure's built-in ingress (or keep Caddy for SSL simplicity)
  - Ensure all services expose correct ports
  - Add health checks for all services
  - Configure persistent volumes for PostgreSQL and MinIO
  - Update environment variable references to use Azure secrets

  **Must NOT do:**
  - Bind mount host directories (use Azure Files for persistence)
  - Hardcode any secrets
  - Use `network_mode: host` (not supported in Container Apps)

  **Recommended Agent Profile:**
  - **Category:** `unspecified-high`
  - **Skills:** None (YAML configuration)
  - **Justification:** Infrastructure-as-code configuration requiring attention to detail

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 2
  - **Blocks:** Task 8 (Deploy)
  - **Blocked By:** Task 2 (Azure Infra - need to know resource names)

  **References:**
  - Current file: `deploy/docker-compose.prod.yml`
  - Azure Container Apps Compose: https://docs.microsoft.com/azure/container-apps/compose

  **Acceptance Criteria:**
  - [ ] File: `deploy/docker-compose.prod.yml` updated
  - [ ] All 5 services defined:
    - `app` (backend) - port 8080
    - `web` (frontend) - port 3000
    - `postgres` - port 5432, persistent volume
    - `redis` - port 6379
    - `minio` - ports 9000, 9001, persistent volume
  - [ ] Health checks configured for all services
  - [ ] Environment variables reference `${VAR_NAME}` (not hardcoded)
  - [ ] No `network_mode: host` usage
  - [ ] Caddy service removed (using Azure ingress) OR updated for Azure

  **Evidence:**
  ```bash
  # Validate YAML syntax
  docker-compose -f deploy/docker-compose.prod.yml config
  ```

  **Commit:** YES
  - Message: `ci(deploy): update docker-compose.prod.yml for Azure Container Apps`
  - Files: `deploy/docker-compose.prod.yml`
  - Pre-commit: Validate YAML syntax

---

- [ ] **5. Create GitHub Actions Workflow**

  **What to do:**
  - Create `.github/workflows/deploy-azure.yml`
  - Trigger: workflow_dispatch (manual) + optional push to main
  - Jobs:
    1. Build and push Docker images
    2. Deploy to Azure Container Apps
  - Use Azure/container-apps-deploy-action for deployment
  - Reference all secrets from GitHub secrets
  - Add environment protection (optional)

  **Must NOT do:**
  - Hardcode Azure credentials
  - Deploy on every PR (security risk)
  - Skip health checks after deployment

  **Recommended Agent Profile:**
  - **Category:** `unspecified-high`
  - **Skills:** `git-master`
  - **Justification:** CI/CD pipeline requiring GitHub Actions expertise

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 2
  - **Blocks:** Task 8 (Deploy)
  - **Blocked By:** Task 3 (GitHub Secrets - need secrets configured)

  **References:**
  - Existing CI: `.github/workflows/ci.yml`
  - Azure Container Apps Action: https://github.com/Azure/container-apps-deploy-action
  - GitHub Actions Docs: https://docs.github.com/actions

  **Acceptance Criteria:**
  - [ ] File: `.github/workflows/deploy-azure.yml` created
  - [ ] Workflow triggers:
    - `workflow_dispatch` (manual button)
    - Optional: `push` to `main` branch
  - [ ] Jobs defined:
    - `build`: Build and push images to Docker Hub
    - `deploy`: Deploy to Azure Container Apps
  - [ ] Uses `AZURE_CREDENTIALS` secret for authentication
  - [ ] Uses `azure/container-apps-deploy-action@v2`
  - [ ] References `deploy/docker-compose.prod.yml`
  - [ ] Environment variables passed from GitHub secrets

  **Evidence:**
  ```bash
  # Validate workflow syntax
  gh workflow view deploy-azure.yml
  # Or check in GitHub UI: Actions tab
  ```

  **Commit:** YES
  - Message: `ci(deploy): add Azure deployment workflow`
  - Files: `.github/workflows/deploy-azure.yml`
  - Pre-commit: Validate YAML syntax with actionlint or similar

---

- [ ] **6. Update Backend Email Configuration**

  **What to do:**
  - Verify backend reads `RESEND_API_KEY` from environment
  - Update `cmd/main.go` to pass Resend config to notification service
  - Ensure `FROM_EMAIL` uses verified domain
  - Remove or update fallback to `onboarding@resend.dev`
  - Add logging for email send attempts

  **Must NOT do:**
  - Hardcode API keys
  - Skip error handling for email failures
  - Remove SMTP fallback (keep as backup)

  **Recommended Agent Profile:**
  - **Category:** `ultrabrain`
  - **Skills:** None (Go code review)
  - **Justification:** Backend logic requiring understanding of email flow

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 2
  - **Blocks:** Task 8 (Deploy)
  - **Blocked By:** Task 1 (Resend Setup - need verified domain)

  **References:**
  - `cmd/main.go:216-242` - Resend client initialization
  - `internal/services/email_service.go` - Email sending logic
  - `internal/services/notification_service.go:101-142` - Resend config

  **Acceptance Criteria:**
  - [ ] `cmd/main.go` reads `RESEND_API_KEY` from environment
  - [ ] `RESEND_FROM_EMAIL` uses verified domain (not `onboarding@resend.dev`)
  - [ ] `RESEND_FROM_NAME` set to "Agromart"
  - [ ] Error handling logs failed attempts
  - [ ] No hardcoded API keys or from-addresses

  **Code Review Points:**
  ```go
  // In cmd/main.go - verify this pattern:
  resendAPIKey := os.Getenv("RESEND_API_KEY")
  resendFromEmail := os.Getenv("RESEND_FROM_EMAIL")
  resendFromName := os.Getenv("RESEND_FROM_NAME")
  if resendAPIKey == "" {
      log.Fatal("RESEND_API_KEY is required")
  }
  ```

  **Commit:** YES
  - Message: `fix(email): update Resend configuration for production`
  - Files: `cmd/main.go` (if changes needed)
  - Pre-commit: `go vet ./cmd/main.go`

---

- [ ] **7. Update Production Environment Files**

  **What to do:**
  - Update `deploy/.env.prod` with production-ready configuration
  - Update `deploy/.env.prod.template` with documentation
  - Ensure all required env vars are documented
  - Add Azure-specific configurations
  - Remove development-only settings (Mailpit, pgAdmin)

  **Must NOT do:**
  - Include real secrets in template files
  - Leave placeholder values that won't work
  - Forget to document new Azure-specific variables

  **Recommended Agent Profile:**
  - **Category:** `quick`
  - **Skills:** None (documentation)
  - **Justification:** Template and documentation updates

  **Parallelization:**
  - **Can Run In Parallel:** YES
  - **Parallel Group:** Wave 2
  - **Blocks:** Task 8 (Deploy)
  - **Blocked By:** Task 1 (Resend Setup - need real values for template)

  **References:**
  - Current: `deploy/.env.prod` and `deploy/.env.prod.template`
  - Example: `.env.example`

  **Acceptance Criteria:**
  - [ ] File: `deploy/.env.prod` updated with Azure configs
  - [ ] File: `deploy/.env.prod.template` updated with documentation
  - [ ] All environment variables documented:
    - Database (PostgreSQL)
    - Cache (Redis)
    - Storage (MinIO)
    - Email (Resend)
    - Security (JWT, CSRF)
    - URLs (FRONTEND_URL, BACKEND_URL)
  - [ ] Placeholder format: `CHANGE_ME_*` for all secrets
  - [ ] Comments explaining each variable

  **Environment Variables Checklist:**
  ```
  # Database
  POSTGRES_USER=agromart
  POSTGRES_PASSWORD=CHANGE_ME_POSTGRES_PASSWORD
  POSTGRES_DB=agromart
  DATABASE_URL=postgresql://agromart:CHANGE_ME_POSTGRES_PASSWORD@postgres:5432/agromart
  
  # Redis
  REDIS_URL=redis://redis:6379
  
  # MinIO
  MINIO_ROOT_USER=minioadmin
  MINIO_ROOT_PASSWORD=CHANGE_ME_MINIO_PASSWORD
  MINIO_ENDPOINT=minio:9000
  MINIO_ACCESS_KEY=minioadmin
  MINIO_SECRET_KEY=CHANGE_ME_MINIO_PASSWORD
  MINIO_USE_SSL=false
  
  # Security
  JWT_SECRET=CHANGE_ME_JWT_SECRET_32_CHARS_MIN
  CSRF_SECRET=CHANGE_ME_CSRF_SECRET_32_CHARS_MIN
  ENFORCE_HTTPS=true
  
  # URLs
  FRONTEND_URL=https://CHANGE_ME.azurecontainerapps.io
  BACKEND_URL=https://CHANGE_ME.azurecontainerapps.io/v1
  
  # Email (Resend)
  RESEND_API_KEY=CHANGE_ME_RESEND_API_KEY
  RESEND_FROM_EMAIL=noreply@CHANGE_ME_DOMAIN.com
  RESEND_FROM_NAME=Agromart
  
  # Razorpay (Optional)
  RAZORPAY_KEY_ID=
  RAZORPAY_KEY_SECRET=
  RAZORPAY_WEBHOOK_SECRET=
  
  # Super Admin
  SUPER_ADMIN_EMAIL=admin@CHANGE_ME_DOMAIN.com
  SUPER_ADMIN_PASSWORD=CHANGE_ME_ADMIN_PASSWORD
  ```

  **Commit:** YES
  - Message: `docs(env): update production environment templates for Azure`
  - Files: `deploy/.env.prod`, `deploy/.env.prod.template`
  - Pre-commit: Verify no real secrets committed

---

### Wave 3: Deployment and Verification

- [ ] **8. Deploy to Azure**

  **What to do:**
  - Trigger GitHub Actions workflow manually
  - Monitor build and deployment progress
  - Verify all containers start successfully
  - Check Azure Container Apps logs for errors
  - Ensure health checks pass

  **Must NOT do:**
  - Skip log review on first deployment
  - Ignore failed health checks
  - Deploy without verifying secrets are accessible

  **Recommended Agent Profile:**
  - **Category:** `unspecified-high`
  - **Skills:** `dev-browser`
  - **Justification:** Deployment execution requiring troubleshooting skills

  **Parallelization:**
  - **Can Run In Parallel:** NO
  - **Parallel Group:** Wave 3 (sequential)
  - **Blocks:** Task 9 (Verification)
  - **Blocked By:** Task 3, 4, 5, 6, 7 (all Wave 2 tasks)

  **References:**
  - GitHub Actions: `.github/workflows/deploy-azure.yml`
  - Azure Portal: Container Apps monitoring
  - Docker Compose: `deploy/docker-compose.prod.yml`

  **Acceptance Criteria:**
  - [ ] GitHub Actions workflow triggered successfully
  - [ ] Docker images built and pushed to registry
  - [ ] Azure deployment completed without errors
  - [ ] All 5 containers show "Running" status in Azure Portal
  - [ ] Health check endpoints return 200 OK:
    - Backend: `https://<backend-url>/health` (or similar)
    - Frontend: `https://<frontend-url>` loads without errors
    - PostgreSQL: Connection successful from backend
    - Redis: Connection successful from backend
    - MinIO: Console accessible (if exposed)
  - [ ] Application logs show no critical errors

  **Evidence:**
  ```bash
  # Check deployment status
  az containerapp show --name agromart-backend --resource-group agromart-rg
  az containerapp show --name agromart-frontend --resource-group agromart-rg
  
  # View logs
  az containerapp logs show --name agromart-backend --resource-group agromart-rg --tail 100
  
  # Test health endpoint
  curl -s https://<backend-url>/health | jq .
  ```

  **Commit:** NO (deployment is external action)

---

- [ ] **9. Verification and Testing**

  **What to do:**
  - Access frontend URL in browser
  - Register a new user account
  - Verify email is received (check inbox and Resend dashboard)
  - Click verification link in email
  - Confirm user status changes to "active"
  - Login with verified credentials
  - Test basic inventory operations
  - Document any issues found

  **Must NOT do:**
  - Consider deployment complete without email verification test
  - Ignore deliverability issues (spam folder, bounce rates)
  - Skip SSL certificate validation

  **Recommended Agent Profile:**
  - **Category:** `unspecified-high`
  - **Skills:** `playwright`, `dev-browser`
  - **Justification:** End-to-end testing requiring browser automation

  **Parallelization:**
  - **Can Run In Parallel:** NO (depends on Task 8)
  - **Parallel Group:** Wave 3
  - **Blocks:** None (final task)
  - **Blocked By:** Task 8 (Deploy)

  **References:**
  - Resend Dashboard: https://resend.com/dashboard
  - Frontend: Deployed Azure URL
  - Backend API: Deployed Azure URL + `/v1`

  **Acceptance Criteria:**
  - [ ] Frontend loads successfully at `https://<azure-url>`
  - [ ] SSL certificate is valid (no browser warnings)
  - [ ] User registration form submits successfully
  - [ ] Verification email received within 5 minutes
  - [ ] Email sender shows as configured from-address (not Resend default)
  - [ ] Verification link in email works and activates account
  - [ ] User can login with activated credentials
  - [ ] Dashboard loads and shows inventory data
  - [ ] Resend dashboard shows "Delivered" status for test email

  **Evidence:**
  ```bash
  # Test frontend accessibility
  curl -s -o /dev/null -w "%{http_code}" https://<frontend-url>
  # Expected: 200
  
  # Test backend API
  curl -s https://<backend-url>/v1/health
  # Expected: {"status":"ok"} or similar
  
  # Verify SSL certificate
  openssl s_client -connect <frontend-url>:443 -servername <frontend-url> 2>/dev/null | openssl x509 -noout -dates
  ```

  **Manual Testing Steps:**
  1. Open browser to `https://<azure-url>`
  2. Click "Register" or "Sign Up"
  3. Fill in: Name, Email (use real email you can access), Password
  4. Submit registration
  5. Check email inbox (and spam folder) for verification email
  6. Open Resend dashboard, confirm email shows as "Delivered"
  7. Click verification link in email
  8. Redirect should go to login page with success message
  9. Login with credentials
  10. Verify dashboard loads with user info

  **Commit:** NO (verification is manual testing)

---

## Environment Variables Required

### GitHub Secrets (Repository Settings)

| Secret Name | Value Source | Required By | Description |
|-------------|--------------|-------------|-------------|
| `RESEND_API_KEY` | Task 1 - Resend Dashboard | Backend, Workflow | Production API key for email sending |
| `AZURE_CREDENTIALS` | Task 2 - Azure CLI | Workflow | Service principal JSON for Azure auth |
| `DOCKER_USERNAME` | Docker Hub | Workflow | Docker Hub username for image push |
| `DOCKER_PASSWORD` | Docker Hub | Workflow | Docker Hub access token |
| `JWT_SECRET` | Generate random | Backend | 32+ character random string for JWT signing |
| `CSRF_SECRET` | Generate random | Backend | 32+ character random string for CSRF protection |
| `POSTGRES_PASSWORD` | Generate random | PostgreSQL | Database password |
| `MINIO_ROOT_PASSWORD` | Generate random | MinIO | Storage admin password |
| `SUPER_ADMIN_PASSWORD` | Generate random | Backend | Initial admin user password |

### Azure Container Apps Secrets

These are configured in Azure and referenced by the containers:

| Secret Name | Mapped To | Description |
|-------------|-----------|-------------|
| `database-url` | `DATABASE_URL` env var | PostgreSQL connection string |
| `redis-url` | `REDIS_URL` env var | Redis connection string |
| `resend-api-key` | `RESEND_API_KEY` env var | Resend API key |
| `jwt-secret` | `JWT_SECRET` env var | JWT signing secret |
| `csrf-secret` | `CSRF_SECRET` env var | CSRF protection secret |
| `minio-root-password` | `MINIO_ROOT_PASSWORD` env var | MinIO admin password |

### Local Environment (For Testing)

Create `deploy/.env.prod` locally (not committed) with real values for manual testing before GitHub deployment.

---

## Commit Strategy

| After Task | Commit Message | Files | Verification |
|------------|----------------|-------|--------------|
| 4 | `ci(deploy): update docker-compose.prod.yml for Azure Container Apps` | `deploy/docker-compose.prod.yml` | `docker-compose -f deploy/docker-compose.prod.yml config` |
| 5 | `ci(deploy): add Azure deployment workflow` | `.github/workflows/deploy-azure.yml` | actionlint or YAML validation |
| 6 | `fix(email): update Resend configuration for production` | `cmd/main.go` (if needed) | `go vet ./cmd/main.go` |
| 7 | `docs(env): update production environment templates for Azure` | `deploy/.env.prod`, `deploy/.env.prod.template` | Manual review |

**Note:** Tasks 1, 2, 3, 8, 9 don't require commits (external configuration or manual testing).

---

## Success Criteria

### Verification Commands

```bash
# 1. Check Azure resources exist
az group show --name agromart-rg
az containerapp env show --name agromart-env --resource-group agromart-rg

# 2. Verify containers are running
az containerapp list --resource-group agromart-rg --output table

# 3. Test frontend accessibility
curl -s -o /dev/null -w "%{http_code}" https://<frontend-url>

# 4. Test backend health endpoint
curl -s https://<backend-url>/v1/health | jq .

# 5. Verify SSL certificate
openssl s_client -connect <frontend-url>:443 -servername <frontend-url> 2>/dev/null | openssl x509 -noout -text | grep "Subject:"

# 6. Check application logs for errors
az containerapp logs show --name agromart-backend --resource-group agromart-rg --tail 50
```

### Final Checklist

- [ ] All Azure resources created (Resource Group, Container Apps Environment)
- [ ] All 5 containers running (frontend, backend, postgres, redis, minio)
- [ ] GitHub Actions workflow executes successfully
- [ ] SSL certificate valid (HTTPS works without warnings)
- [ ] User registration sends real email via Resend
- [ ] Email verification link activates user account
- [ ] User can login and access dashboard
- [ ] No critical errors in application logs
- [ ] Resend dashboard shows email as "Delivered"

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Resend domain verification delays | Medium | High | Start Task 1 immediately; use Resend testing domain as fallback |
| Azure service principal permission issues | Low | High | Use contributor role; test with `az login --service-principal` |
| Database persistence loss on restart | Medium | Medium | Configure Azure Files volume for PostgreSQL; document it's for demo only |
| Secrets exposed in logs | Low | Critical | Use GitHub Secrets masking; verify no `echo $SECRET` in workflow |
| Email deliverability issues (spam) | Medium | High | Verify domain with SPF/DKIM; test with multiple email providers |
| SSL certificate provisioning delays | Low | Medium | Use Azure-managed certs; fallback to Let's Encrypt via Caddy if needed |
| Container startup failures | Medium | High | Add health checks; review logs immediately after deployment |

---

## Post-Deployment Recommendations

1. **Monitoring:** Set up Azure Monitor alerts for container restarts
2. **Backups:** For production use, migrate to Azure Database for PostgreSQL with automated backups
3. **Scaling:** Configure Azure Container Apps scaling rules based on CPU/memory
4. **Security:** Enable Azure Key Vault for secret management (instead of Container Apps secrets)
5. **CDN:** Add Azure CDN in front of frontend for global performance
6. **Custom Domain:** Configure custom domain with Azure DNS for production

---

## Troubleshooting Guide

### Common Issues

**Email not sending:**
- Check Resend dashboard for "Bounced" or "Dropped" status
- Verify `RESEND_API_KEY` is set correctly in Azure secrets
- Check backend logs for Resend API errors
- Verify `RESEND_FROM_EMAIL` matches verified domain

**Containers failing to start:**
- Check Azure Container Apps logs: `az containerapp logs show`
- Verify environment variables are set in Azure secrets
- Check Docker Compose syntax: `docker-compose config`
- Ensure all required secrets are created

**Database connection errors:**
- Verify `DATABASE_URL` format includes correct service name
- Check PostgreSQL container health
- Ensure database migrations have run

**SSL/HTTPS issues:**
- Verify Azure Container Apps ingress is configured for HTTPS
- Check certificate provisioning status in Azure Portal
- Ensure `ENFORCE_HTTPS=true` in backend config

---

## Resources and References

### Documentation
- [Azure Container Apps Documentation](https://docs.microsoft.com/azure/container-apps/)
- [Resend Documentation](https://resend.com/docs)
- [GitHub Actions Documentation](https://docs.github.com/actions)
- [Docker Compose Specification](https://docs.docker.com/compose/compose-file/)

### Azure CLI Commands Reference
```bash
# Resource Group
az group create --name agromart-rg --location eastus

# Container Apps Environment
az containerapp env create --name agromart-env --resource-group agromart-rg --location eastus

# Service Principal
az ad sp create-for-rbac --name "agromart-github-actions" --role contributor --scopes /subscriptions/{sub-id}/resourceGroups/agromart-rg --sdk-auth

# Deploy Container Apps
az containerapp create --name agromart-backend --resource-group agromart-rg --environment agromart-env --image {docker-image} --target-port 8080
```

---

**Plan Generated:** 2026-01-30
**Total Tasks:** 9
**Estimated Duration:** 6-8 hours
**Parallelizable Waves:** 3
