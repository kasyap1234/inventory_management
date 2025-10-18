# 🚀 Quick Start: Subscription System Setup

## ✅ What's Already Done

Your subscription system is **fully implemented** and ready to use! Here's what's working:

### Backend ✅
- Razorpay API integration
- Subscription CRUD operations
- Webhook handling
- 3 predefined plans (Basic, Premium, Enterprise)
- Database schema and repositories
- Security and error handling

### Frontend ✅
- Plan selection page (`/dashboard/subscriptions/plans`)
- Subscription management page (`/dashboard/subscriptions`)
- Beautiful, responsive UI
- Real-time status updates

---

## 🔧 Setup Steps (5 minutes)

### Step 1: Get Razorpay Credentials

1. **Sign up**: https://razorpay.com/
2. **Get API Keys**:
   - Dashboard → Settings → API Keys
   - Generate **Test Mode** keys first
   - Copy `Key ID` and `Key Secret`

### Step 2: Configure Environment Variables

Add to your `.env` file:

```bash
# Razorpay Test Keys (for development)
RAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxxx
RAZORPAY_KEY_SECRET=your_test_secret_key
RAZORPAY_WEBHOOK_SECRET=your_webhook_secret
```

### Step 3: Setup Webhook (Optional for local testing)

For local development, use **ngrok**:

```bash
# Install ngrok
brew install ngrok  # macOS

# Expose your local server
ngrok http 8080

# Copy the HTTPS URL and add to Razorpay:
# Dashboard → Settings → Webhooks
# URL: https://abc123.ngrok.io/v1/webhooks/razorpay
# Events: Select all subscription.* events
```

### Step 4: Start Your Application

```bash
# Terminal 1: Start Backend
cd /path/to/inventory_management
go run cmd/main.go

# Terminal 2: Start Frontend
cd frontend
npm run dev
```

### Step 5: Test Subscription

1. Open browser: `http://localhost:3000`
2. Login to your account
3. Navigate to: **Subscriptions** → **View Plans**
4. Click **Subscribe Now** on any plan
5. ✅ Subscription created!

---

## 📋 Available Plans

| Plan | Price | Features |
|------|-------|----------|
| **Basic** | ₹999/month | 5 warehouses, Basic features |
| **Premium** | ₹2,499/month | 20 warehouses, Advanced analytics |
| **Enterprise** | ₹4,999/month | Unlimited, Full features |

---

## 🧪 Test Scenarios

### Test 1: Create Subscription
```
1. Go to /dashboard/subscriptions/plans
2. Click "Subscribe Now" on Premium plan
3. Check /dashboard/subscriptions
4. ✅ Should see new subscription with status "created"
```

### Test 2: Pause Subscription
```
1. Go to /dashboard/subscriptions
2. Click "Pause" on active subscription
3. ✅ Status changes to "paused"
```

### Test 3: Resume Subscription
```
1. Click "Resume" on paused subscription
2. ✅ Status changes to "active"
```

### Test 4: Cancel Subscription
```
1. Click "Cancel" on active subscription
2. ✅ Status changes to "cancelled"
```

---

## 🔍 Verify Everything Works

### Check Backend
```bash
# Check if Razorpay service is initialized
curl http://localhost:8080/health

# List available plans
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/v1/subscriptions/plans
```

### Check Database
```sql
-- View subscriptions
SELECT * FROM subscriptions ORDER BY created_at DESC;

-- Check plans are available
-- Plans are hardcoded in backend (basic, premium, enterprise)
```

### Check Frontend
```
✅ /dashboard/subscriptions/plans - Plan selection page
✅ /dashboard/subscriptions - Subscription management
✅ Sidebar has "Subscriptions" link
```

---

## 🐛 Common Issues

### Issue: "Razorpay API error"
**Solution**: Check your API keys in `.env`

### Issue: Webhook not working
**Solution**: 
- For local: Use ngrok
- For production: Use HTTPS URL
- Verify webhook secret matches

### Issue: Subscription not showing
**Solution**: Check database connection and JWT token

---

## 📚 Documentation

Full documentation: `docs/SUBSCRIPTION_SYSTEM.md`

Includes:
- Complete API reference
- Webhook setup guide
- Security best practices
- Production deployment
- Troubleshooting guide

---

## 🎉 You're Ready!

Your subscription system is **production-ready**. You can:

✅ Accept real payments (switch to live keys)  
✅ Manage subscriptions  
✅ Handle webhooks automatically  
✅ Scale to multiple tenants  

**Start accepting subscriptions now!** 🚀

---

## 🆘 Need Help?

1. Check logs: Backend console or `/var/log/agromart/`
2. Review Razorpay dashboard for payment status
3. Read full docs: `docs/SUBSCRIPTION_SYSTEM.md`
4. Test with Razorpay test cards: https://razorpay.com/docs/payments/payments/test-card-details/
