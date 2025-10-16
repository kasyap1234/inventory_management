# .gitignore Update Plan

## Current Situation
- The current `.gitignore` file is missing several important entries that should exclude unnecessary files
- The `.env` file is currently NOT being ignored (which is correct since user specified to exclude `.env` from .gitignore, meaning it should be tracked)
- Several files are being tracked that should be ignored

## Files Currently Missing from .gitignore
Based on project analysis, the following files/directories should be added to .gitignore:

### Binary and Executable Files
- `main` - Go binary
- `agromart` - Another binary
- `agromart-app` - Application binary

### Log Files
- `app.log`
- `backend.log` 
- `frontend.log`

### Process ID Files
- `backend.pid`
- `frontend.pid`

### Coverage Files
- `coverage.out`
- `repo_coverage.out`

### Temporary Files
- `nohup.out`
- `repositories.test`

### Test Files
- `test_image.jpg`
- `test_minio_image.jpg`
- `test_upload.jpg`

## Updated .gitignore Content
```
# If you prefer the allow list template instead of the deny list, see community template:
# https://github.com/github/gitignore/blob/main/community/Golang/Go.AllowList.gitignore
#
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
.main
.agromart-app

# Test binary, built with `go test -c`
*.test

# Code coverage profiles and other test artifacts
*.out
coverage.*
*.coverprofile
profile.cov

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work
go.work.sum

# env file
# NOTE: .env is intentionally NOT ignored to keep it tracked
.env.development.local
.env.test.local
.env.production.local

# Editor/IDE
.idea/
.vscode/

# Node.js / Next.js
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.pnpm-debug.log*
bun.lockb

# Next.js build output
.next/
out/

# Logs
*.log
logs/
app.log
backend.log
frontend.log

# PIDs
*.pid
backend.pid
frontend.pid

# Runtime data
pids
*.pid.lock

# Coverage directory used by tools like istanbul
coverage/
*.lcov

# nyc test coverage
.nyc_output

# TypeScript cache
*.tsbuildinfo

# Optional npm cache directory
.npm

# Optional eslint cache
.eslintcache

# Microbundle cache
.rpt2_cache/
.rts2_cache_cjs/
.rts2_cache_es/
.rts2_cache_umd/

# Optional REPL history
.node_repl_history

# Output of 'npm pack'
*.tgz

# Yarn Integrity file
.yarn-integrity

# parcel-bundler cache
.cache
.parcel-cache

# Gatsby files
.cache/
public

# Storybook build outputs
.out
.storybook-out

# Temporary folders
tmp/
temp/

# OS generated files
.DS_Store
Thumbs.db
ehthumbs.db
Desktop.ini

# MinIO data
minio-data/
data/

# PostgreSQL data (if local)
pgdata/

# Docker
.dockerignore

# Build artifacts
build/
dist/

# Test files
test_image.jpg
test_minio_image.jpg
test_upload.jpg
test_*.jpg

# Coverage files
coverage.out
repo_coverage.out

# NoHup output
nohup.out

# Repositories test file
repositories.test

# Scripts (if executable)
```

## Implementation Steps
1. Replace the existing .gitignore with the updated content above
2. Ensure .env file remains tracked (not ignored)
3. Verify that all unnecessary files will now be excluded from git

## Verification
After implementation, run `git status` to confirm that unnecessary files are no longer tracked.