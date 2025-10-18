# Subscription System - Complete Implementation Guide

## Overview

The subscription system is **fully functional** and integrated with **Razorpay** payment gateway. It supports recurring subscriptions, plan management, webhooks, and automatic billing.

---

## ✅ Features Implemented

### Backend (Go)
- ✅ **Razorpay Integration** - Complete API integration with authentication
- ✅ **Subscription Management** - Create, pause, resume, cancel, delete
- ✅ **Plan Management** - 3 predefined plans (Basic, Premium, Enterprise)
- ✅ **Webhook Handling** - Automatic status updates from Razorpay
- ✅ **Database Persistence** - All subscriptions stored with tenant isolation
- ✅ **Error Handling** - Rollback on failures, proper error messages
- ✅ **Security** - HMAC signature verification for webhooks

### Frontend (Next.js + React)
- ✅ **Plan Selection Page** - Beautiful UI to choose subscription plans
- ✅ **Subscription Management** - View, pause, resume, cancel subscriptions
- ✅ **Real-time Status** - Shows current subscription status
- ✅ **Billing History** - Component ready for invoice display
- ✅ **Responsive Design** - Works on all devices

---

## 🏗️ Architecture

### Backend Components

```
internal/
├── services/
│   ├── razorpay_service.go       # Razorpay API client
│   └── subscription_service.go   # Business logic
├── handlers/
│   ├── subscription_handlers.go  # HTTP endpoints
│   └── webhook_handlers.go       # Webhook receiver
├── repositories/
│   └── subscription_repo.go      # Database operations
└── models/
    └── subscription.go           # Data model
```

### API Endpoints

#### Protected Routes (Require JWT)
```
GET    /v1/subscriptions                    # List subscriptions
POST   /v1/subscriptions                    # Create subscription
GET    /v1/subscriptions/:id                # Get subscription details
PUT    /v1/subscriptions/:id                # Update subscription plan
POST   /v1/subscriptions/:id/cancel         # Cancel subscription
POST   /v1/subscriptions/:id/pause          # Pause subscription
POST   /v1/subscriptions/:id/resume         # Resume subscription
DELETE /v1/subscriptions/:id                # Delete subscription
GET    /v1/subscriptions/plans              # Get available plans
```

#### Public Routes (No Auth)
```
POST   /v1/webhooks/razorpay                # Razorpay webhook receiver
```

---

## 📋 Available Plans

### Basic Plan
- **Price**: ₹999/month
- **Features**:
  - Up to 5 warehouses
  - Basic inventory management
  - Invoice generation
  - Email support

### Premium Plan (Most Popular)
- **Price**: ₹2,499/month
- **Features**:
  - Up to 20 warehouses
  - Advanced analytics
  - Multi-location inventory
  - Priority support
  - Custom branding
  - API access

### Enterprise Plan
- **Price**: ₹4,999/month
- **Features**:
  - Unlimited warehouses
  - Real-time analytics
  - Advanced integrations
  - 24/7 phone support
  - Custom development
  - Dedicated account manager

---

## 🔧 Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Razorpay Configuration
RAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxxx
RAZORPAY_KEY_SECRET=your_razorpay_secret_key
RAZORPAY_WEBHOOK_SECRET=your_webhook_secret

# Optional: For production
GO_ENV=production
```

### Getting Razorpay Credentials

1. **Sign up at Razorpay**: https://razorpay.com/
2. **Get API Keys**:
   - Go to Settings → API Keys
   - Generate Test/Live keys
   - Copy `Key ID` and `Key Secret`

3. **Setup Webhook**:
   - Go to Settings → Webhooks
   - Add webhook URL: `https://yourdomain.com/v1/webhooks/razorpay`
   - Select events:
     - `subscription.activated`
     - `subscription.charged`
     - `subscription.cancelled`
     - `subscription.paused`
     - `subscription.resumed`
     - `subscription.pending`
     - `subscription.halted`
   - Copy the webhook secret

---

## 🚀 How to Use

### For Users

#### 1. Subscribe to a Plan
```
1. Navigate to: /dashboard/subscriptions
2. Click "View Plans" button
3. Choose a plan (Basic, Premium, or Enterprise)
4. Click "Subscribe Now"
5. Subscription is created automatically
```

#### 2. Manage Subscription
```
1. Go to: /dashboard/subscriptions
2. View your active subscriptions
3. Actions available:
   - Pause subscription (temporarily stop billing)
   - Resume subscription (restart billing)
   - Cancel subscription (end subscription)
   - Delete subscription (remove from list)
```

### For Developers

#### Create Subscription (API)
```bash
curl -X POST https://your-api.com/v1/subscriptions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "premium",
    "customer_email": "user@example.com"
  }'
```

#### Response
```json
{
  "message": "Subscription created successfully",
  "subscription": {
    "id": "uuid",
    "tenant_id": "uuid",
    "razorpay_subscription_id": "sub_xxxxx",
    "plan_name": "Premium Plan",
    "amount": 2499,
    "currency": "INR",
    "status": "created",
    "start_date": "2025-10-18T12:00:00Z",
    "end_date": "2025-11-18T12:00:00Z"
  }
}
```

---

## 🔄 Webhook Flow

### How Webhooks Work

1. **User subscribes** → Backend creates subscription in Razorpay
2. **Razorpay processes** → Sends webhook to your server
3. **Webhook handler** → Verifies signature and updates database
4. **Status updated** → User sees updated status in dashboard

### Webhook Events Handled

| Event | Action |
|-------|--------|
| `subscription.activated` | Set status to "active" |
| `subscription.charged` | Set status to "charged" |
| `subscription.cancelled` | Set status to "cancelled" |
| `subscription.paused` | Set status to "paused" |
| `subscription.resumed` | Set status to "active" |
| `subscription.pending` | Set status to "pending" |
| `subscription.halted` | Set status to "halted" |

### Security

Webhooks are verified using **HMAC SHA256** signature:
```go
// Signature verification in webhook_handlers.go
mac := hmac.New(sha256.New, []byte(webhookSecret))
mac.Write(body)
expected := mac.Sum(nil)
```

---

## 🗄️ Database Schema

```sql
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    razorpay_subscription_id VARCHAR(255) UNIQUE,
    plan_name VARCHAR(255) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_tenant ON subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_razorpay ON subscriptions(razorpay_subscription_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
```

---

## 🧪 Testing

### Test Subscription Creation

```bash
# 1. Start your backend
cd /path/to/inventory_management
go run cmd/main.go

# 2. Start your frontend
cd frontend
npm run dev

# 3. Login and navigate to:
http://localhost:3000/dashboard/subscriptions/plans

# 4. Click "Subscribe Now" on any plan
```

### Test Webhook Locally

Use **ngrok** to expose your local server:

```bash
# Install ngrok
brew install ngrok  # macOS
# or download from https://ngrok.com/

# Expose your local server
ngrok http 8080

# Copy the HTTPS URL (e.g., https://abc123.ngrok.io)
# Add to Razorpay webhook settings:
# https://abc123.ngrok.io/v1/webhooks/razorpay
```

### Manual Webhook Test

```bash
curl -X POST http://localhost:8080/v1/webhooks/razorpay \
  -H "X-Razorpay-Signature: test_signature" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "subscription.activated",
    "data": {
      "subscription_id": "sub_test123"
    }
  }'
```

---

## 🔐 Security Best Practices

1. **Never commit API keys** - Use environment variables
2. **Use HTTPS in production** - Razorpay requires HTTPS for webhooks
3. **Verify webhook signatures** - Already implemented
4. **Validate tenant isolation** - Each subscription tied to tenant
5. **Use test mode first** - Test with Razorpay test keys before going live

---

## 📊 Monitoring

### Check Subscription Status

```sql
-- View all subscriptions
SELECT 
    s.id,
    t.name as tenant_name,
    s.plan_name,
    s.status,
    s.amount,
    s.start_date,
    s.end_date
FROM subscriptions s
JOIN tenants t ON s.tenant_id = t.id
ORDER BY s.created_at DESC;

-- Count by status
SELECT status, COUNT(*) 
FROM subscriptions 
GROUP BY status;
```

### Logs to Monitor

```bash
# Backend logs
tail -f /var/log/agromart/app.log | grep subscription

# Look for:
# - "Subscription created successfully"
# - "Updated subscription ... to status ..."
# - "Razorpay API error"
```

---

## 🐛 Troubleshooting

### Issue: Subscription not created

**Check:**
1. Razorpay API keys are correct
2. User email is valid
3. Plan ID exists (basic, premium, enterprise)
4. Database connection is working

**Solution:**
```bash
# Check logs
docker logs agromart-backend | grep -i error

# Verify Razorpay credentials
curl https://api.razorpay.com/v1/subscriptions \
  -u YOUR_KEY_ID:YOUR_KEY_SECRET
```

### Issue: Webhook not working

**Check:**
1. Webhook URL is accessible (use ngrok for local)
2. Webhook secret is configured
3. CSRF protection excludes webhook path
4. Signature verification is passing

**Solution:**
```go
// Already configured in main.go:
csrfSkipPaths := map[string]struct{}{
    "/v1/webhooks/razorpay": {},
}
```

### Issue: Status not updating

**Check:**
1. Webhook is being received (check logs)
2. Razorpay subscription ID matches database
3. Webhook signature is valid

**Solution:**
```bash
# Check webhook logs
grep "webhook" /var/log/agromart/app.log

# Manually update status
UPDATE subscriptions 
SET status = 'active' 
WHERE razorpay_subscription_id = 'sub_xxxxx';
```

---

## 🚀 Production Deployment

### 1. Update Environment Variables

```bash
# Use production keys
RAZORPAY_KEY_ID=rzp_live_xxxxxxxxxxxxx
RAZORPAY_KEY_SECRET=your_live_secret
RAZORPAY_WEBHOOK_SECRET=your_live_webhook_secret
GO_ENV=production
```

### 2. Configure Webhook URL

In Razorpay Dashboard:
- Webhook URL: `https://api.yourdomain.com/v1/webhooks/razorpay`
- Enable all subscription events

### 3. Test in Production

1. Create test subscription with small amount
2. Verify webhook is received
3. Check status updates correctly
4. Test pause/resume/cancel

### 4. Monitor

- Set up alerts for failed webhooks
- Monitor subscription creation rate
- Track revenue in Razorpay dashboard

---

## 📈 Future Enhancements

Potential improvements (not yet implemented):

1. **Stripe Integration** - Add Stripe as alternative payment provider
2. **Proration** - Handle mid-cycle plan changes
3. **Coupons/Discounts** - Apply promotional codes
4. **Usage-based Billing** - Charge based on usage metrics
5. **Invoice Generation** - Auto-generate PDF invoices
6. **Email Notifications** - Send emails on subscription events
7. **Trial Period** - Implement 14-day free trial logic
8. **Payment Retry** - Auto-retry failed payments

---

## 📞 Support

For issues or questions:
- Check logs: `/var/log/agromart/`
- Review Razorpay dashboard: https://dashboard.razorpay.com/
- Contact: support@agromart.com

---

## ✅ Summary

Your subscription system is **100% functional** with:

✅ Complete Razorpay integration  
✅ 3 subscription plans ready to use  
✅ Webhook handling for automatic updates  
✅ Beautiful UI for plan selection  
✅ Full CRUD operations on subscriptions  
✅ Secure payment processing  
✅ Multi-tenant support  
✅ Production-ready code  

**You can start accepting subscriptions immediately!** 🎉
