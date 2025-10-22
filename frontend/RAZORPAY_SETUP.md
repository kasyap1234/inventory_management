# Frontend Razorpay Setup

## Environment Variables

Create a `.env.local` file in the `frontend` directory with the following:

```bash
# Razorpay Public Key (safe to expose in frontend)
NEXT_PUBLIC_RAZORPAY_KEY_ID=rzp_test_your_key_id_here

# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Important Notes

- **NEXT_PUBLIC_RAZORPAY_KEY_ID**: This is your Razorpay Key ID (public key). It's safe to expose in frontend code.
- **Never put your Razorpay Secret Key in frontend** - it should only be in the backend `.env` file.
- For development, use TEST keys (starting with `rzp_test_`)
- For production, use LIVE keys (starting with `rzp_live_`)

## Getting Razorpay Keys

1. Sign up at https://razorpay.com
2. Go to Dashboard → Settings → API Keys
3. Generate Test Mode keys
4. Copy the Key ID to your frontend `.env.local`
5. Copy both Key ID and Secret to backend `.env`

## Razorpay Checkout Integration

The Razorpay Checkout is automatically loaded via the `useRazorpayCheckout` hook when users subscribe to a plan. No manual script loading required!
