import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';

// Razorpay script loader
const loadRazorpayScript = (): Promise<boolean> => {
  return new Promise((resolve) => {
    // Check if script is already loaded
    if (typeof window !== 'undefined' && (window as any).Razorpay) {
      resolve(true);
      return;
    }

    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => resolve(true);
    script.onerror = () => resolve(false);
    document.body.appendChild(script);
  });
};

interface RazorpayOptions {
  key: string;
  amount: number;
  currency: string;
  name: string;
  description: string;
  order_id?: string;
  subscription_id?: string;
  prefill?: {
    name?: string;
    email?: string;
    contact?: string;
  };
  theme?: {
    color?: string;
  };
  handler: (response: any) => void;
  modal?: {
    ondismiss?: () => void;
  };
}

export const useRazorpayCheckout = () => {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [isScriptLoaded, setIsScriptLoaded] = useState(false);
  const [publicKey, setPublicKey] = useState<string | null>(null);

  useEffect(() => {
    loadRazorpayScript().then((loaded) => {
      setIsScriptLoaded(loaded);
    });
    // Fetch Razorpay key from backend config to avoid leaking secrets in code
    api
      .get('/payments/config')
      .then((res) => {
        setPublicKey(res.data?.razorpay_key_id || null);
      })
      .catch(() => {
        // Ignore - will fall back to env
      });
  }, []);

  const openCheckout = async (options: RazorpayOptions) => {
    if (!isScriptLoaded) {
      const loaded = await loadRazorpayScript();
      if (!loaded) {
        console.error('Razorpay SDK failed to load');
        return;
      }
      setIsScriptLoaded(true);
    }

    setIsLoading(true);

    const rzp = new (window as any).Razorpay(options);
    rzp.open();

    setIsLoading(false);
  };

  const resolveKey = () => {
    return publicKey || process.env.NEXT_PUBLIC_RAZORPAY_KEY_ID || 'rzp_test_key';
  };

  const createSubscription = async (
    planId: string,
    customerEmail: string,
    onSuccess?: (subscriptionId: string) => void,
    onError?: (error: any) => void
  ) => {
    try {
      setIsLoading(true);

      // Call backend API to create subscription
      const response = await api.post('/subscriptions', {
        plan_id: planId,
        customer_email: customerEmail,
      });
      const subscription = response.data?.subscription;

      // Open Razorpay Checkout
      await openCheckout({
        key: resolveKey(),
        amount: subscription.amount * 100, // Convert to paise
        currency: subscription.currency || 'INR',
        name: 'AgroMart Subscription',
        description: subscription.plan_name,
        subscription_id: subscription.razorpay_subscription_id || '',
        prefill: {
          email: customerEmail,
        },
        theme: {
          color: '#10b981', // Tailwind green-500
        },
        handler: async (response: any) => {
          // Payment successful
          console.log('Payment successful:', response);
          
          // Optionally verify payment on backend
          try {
            onSuccess?.(subscription.id);
            router.push('/dashboard/subscriptions?success=true');
          } catch (err) {
            console.error('Payment verification failed:', err);
            onError?.(err);
          }
        },
        modal: {
          ondismiss: () => {
            console.log('Checkout dismissed');
            setIsLoading(false);
          },
        },
      });
    } catch (error) {
      console.error('Subscription creation failed:', error);
      onError?.(error);
      setIsLoading(false);
    }
  };

  const createOneTimePayment = async ({
    amount,
    currency = 'INR',
    receipt,
    orderId,
    notes,
    onSuccess,
    onError,
  }: {
    amount: number;
    currency?: string;
    receipt?: string;
    orderId?: string;
    notes?: Record<string, string>;
    onSuccess?: () => void;
    onError?: (error: any) => void;
  }) => {
    try {
      setIsLoading(true);
      const response = await api.post('/payments/orders', {
        amount,
        currency,
        receipt,
        order_id: orderId,
        notes,
      });

      const razorpay = response.data?.razorpay;
      if (!razorpay?.order_id) {
        throw new Error('Failed to create Razorpay order');
      }

      await openCheckout({
        key: resolveKey(),
        order_id: razorpay.order_id,
        amount: razorpay.amount,
        currency: razorpay.currency || currency,
        name: 'AgroMart Payment',
        description: receipt || 'Order payment',
        handler: async (resp: any) => {
          try {
            await api.post('/payments/verify', {
              razorpay_order_id: razorpay.order_id,
              razorpay_payment_id: resp.razorpay_payment_id,
              razorpay_signature: resp.razorpay_signature,
            });
            onSuccess?.();
          } catch (err) {
            console.error('Payment verification failed:', err);
            onError?.(err);
          }
        },
      });
    } catch (error) {
      console.error('Payment creation failed:', error);
      onError?.(error);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    isScriptLoaded,
    openCheckout,
    createSubscription,
    createOneTimePayment,
  };
};
