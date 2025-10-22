import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

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

  useEffect(() => {
    loadRazorpayScript().then((loaded) => {
      setIsScriptLoaded(loaded);
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

  const createSubscription = async (
    planId: string,
    customerEmail: string,
    onSuccess?: (subscriptionId: string) => void,
    onError?: (error: any) => void
  ) => {
    try {
      setIsLoading(true);

      // Call your backend API to create subscription
      const response = await fetch('/api/subscriptions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          plan_id: planId,
          customer_email: customerEmail,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to create subscription');
      }

      const data = await response.json();
      const subscription = data.subscription;

      // Get Razorpay Key from environment or backend
      const razorpayKey = process.env.NEXT_PUBLIC_RAZORPAY_KEY_ID || 'rzp_test_key';

      // Open Razorpay Checkout
      await openCheckout({
        key: razorpayKey,
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
            const verifyResponse = await fetch('/api/subscriptions/verify', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
              },
              body: JSON.stringify({
                razorpay_payment_id: response.razorpay_payment_id,
                razorpay_subscription_id: response.razorpay_subscription_id,
                razorpay_signature: response.razorpay_signature,
              }),
            });

            if (verifyResponse.ok) {
              onSuccess?.(subscription.id);
              router.push('/dashboard/subscriptions?success=true');
            }
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

  return {
    isLoading,
    isScriptLoaded,
    openCheckout,
    createSubscription,
  };
};
