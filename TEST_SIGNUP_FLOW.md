# Account Creation Issue - Diagnosis and Solution

**Date:** January 11, 2025  
**Issue:** Internal Server Error when creating account

---

## 🔍 Problem Identified

The error message in logs shows:
```
Failed to create user kasyap3103@gmail.com for tenant [...]: user with email 'kasyap3103@gmail.com' already exists
```

**Root Cause:** The user account already exists in the database but is in `pending_verification` status.

When trying to login:
```
Email not verified. Please check your inbox for the verification link.
```

---

## ✅ Services Status

All required services are running properly:
- ✅ Backend API: http://localhost:8081 (HEALTHY)
- ✅ Frontend: http://localhost:3000 (RUNNING)
- ✅ PostgreSQL: localhost:5440 (HEALTHY)
- ✅ Redis: localhost:6379 (HEALTHY)
- ✅ MinIO: localhost:9003 (HEALTHY)
- ✅ MailHog: http://localhost:8025 (RUNNING)

---

## 🛠️ Solutions

### Solution 1: Verify Email (RECOMMENDED)

1. **Open MailHog UI:** http://localhost:8025
2. **Find the verification email** for kasyap3103@gmail.com
3. **Click the verification link** in the email
4. **Try logging in again**

### Solution 2: Delete and Recreate Account

If you want to start fresh with a different email:

```bash
# Delete the existing user
psql "postgresql://testuser:testpass@localhost:5440/testdb" -c \
  "DELETE FROM users WHERE email = 'kasyap3103@gmail.com';"

# Now sign up with the same or different email
```

### Solution 3: Manually Verify User (For Testing)

```bash
# Update user status to active
psql "postgresql://testuser:testpass@localhost:5440/testdb" -c \
  "UPDATE users SET status = 'active' WHERE email = 'kasyap3103@gmail.com';"
```

---

## 🧪 Complete End-to-End Test Flow

### Step 1: Sign Up
```bash
# Test signup API
curl -X POST http://localhost:8081/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Expected Response:** 201 Created
```json
{
  "message": "User created successfully. Please check your email to verify your account.",
  "user": {
    "id": "...",
    "email": "test@example.com",
    "status": "pending_verification"
  }
}
```

### Step 2: Check Email
1. Open MailHog: http://localhost:8025
2. Find verification email
3. Click verification link

### Step 3: Login
```bash
# Test login API
curl -X POST http://localhost:8081/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

**Expected Response:** 200 OK with tokens
```json
{
  "access_token": "eyJhbG...",
  "refresh_token": "eyJhbG...",
  "user": {
    "id": "...",
    "email": "test@example.com",
    "status": "active"
  }
}
```

### Step 4: Access Protected Routes
```bash
# Get CSRF token first
CSRF_TOKEN=$(curl -s http://localhost:8081/v1/security/csrf | jq -r '.csrf_token')

# Access protected route
curl -X GET http://localhost:8081/v1/products \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN"
```

---

## 🌐 Frontend Testing

### 1. Open Frontend
Navigate to: http://localhost:3000

### 2. Sign Up Flow
1. Click "Sign Up" or navigate to `/signup`
2. Fill in the form:
   - Email: testnew@example.com
   - Password: SecurePass123!
   - First Name: Test
   - Last Name: User
3. Click "Create Account"
4. **Check MailHog** for verification email

### 3. Verify Email
1. Open http://localhost:8025
2. Click on the email
3. Click "Verify Email" button
4. Should redirect to login page

### 4. Login
1. Navigate to `/login`
2. Enter credentials
3. Should redirect to dashboard

### 5. Test Dashboard Features
- ✅ View dashboard statistics
- ✅ Navigate to products page
- ✅ Create a product
- ✅ View inventory
- ✅ Create an order
- ✅ Generate invoice

---

## 🐛 Common Issues & Fixes

### Issue: "Email not verified"
**Solution:** Check MailHog at http://localhost:8025 for verification email

### Issue: "User already exists"
**Solution:** Either:
1. Login with existing account after verifying email
2. Use a different email address
3. Delete existing user from database

### Issue: "Failed to send email"
**Solution:** 
1. Check if MailHog is running: `docker ps | grep mailhog`
2. Check backend can connect to SMTP: port 1025

### Issue: "Tenant not found"
**Solution:** This is normal - tenant is created automatically during signup

### Issue: Backend not responding
**Solution:**
```bash
# Check if backend is running
ps aux | grep "./main"

# Check backend logs
tail -f backend.log

# Restart if needed
./run_stop.sh
./run_start.sh
```

---

## 📊 Quick Database Queries

### Check all users
```sql
SELECT id, email, status, created_at 
FROM users 
ORDER BY created_at DESC 
LIMIT 10;
```

### Check user roles
```sql
SELECT u.email, r.name as role_name
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'your@email.com';
```

### Check tenants
```sql
SELECT id, name, subdomain, status 
FROM tenants 
ORDER BY created_at DESC;
```

### View recent audit logs
```sql
SELECT user_id, action, entity_type, created_at
FROM audit_logs
ORDER BY created_at DESC
LIMIT 20;
```

---

## 🔧 Reset Everything (Nuclear Option)

If you want to completely start fresh:

```bash
# Stop all services
./run_stop.sh

# Remove all data
docker compose down -v

# Start everything fresh
./run_start.sh
```

This will:
- Stop all containers
- Delete all data volumes
- Recreate everything from scratch
- Run migrations
- Give you a clean slate

---

## ✅ Verification Checklist

Before reporting issues, verify:

- [ ] All Docker containers are running (postgres, redis, minio, mailhog)
- [ ] Backend is accessible at http://localhost:8081/health
- [ ] Frontend is accessible at http://localhost:3000
- [ ] MailHog UI loads at http://localhost:8025
- [ ] Database connection works (psql command)
- [ ] No errors in backend.log
- [ ] No errors in frontend.log

---

## 📝 Test Results

**Backend Health:** ✅ HEALTHY
```json
{
  "service": "agromart2",
  "status": "ok"
}
```

**Frontend:** ✅ RUNNING on port 3000

**Database:** ✅ CONNECTED (PostgreSQL 17.5)

**Services:** ✅ ALL RUNNING
- PostgreSQL: 5440
- Redis: 6379
- MinIO: 9003/9004
- MailHog: 1025/8025

---

## 🎯 Next Steps

1. **For existing user (kasyap3103@gmail.com):**
   - Go to http://localhost:8025
   - Find and click verification email
   - Then try logging in

2. **For new testing:**
   - Use a different email address
   - Or delete existing user first
   - Follow signup flow again

3. **To verify the fix:**
   - Try the complete E2E flow with a new email
   - Document any issues found
   - Update this test document

---

**Status:** Issue Diagnosed ✅  
**Solution:** Email verification required  
**Action:** Check MailHog at http://localhost:8025
