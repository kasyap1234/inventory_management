import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { loadStripe, Stripe } from '@stripe/stripe-js';

// Stripe publishable key from environment
const stripePromise = loadStripe(
  process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || ''
);

interface StripeCheckoutOptions {
  priceId: string;
  customerEmail: string;
  successUrl?: string;
  cancelUrl?: string;
  mode?: 'subscription' | 'payment';
  metadata?: Record<string, string>;
}

export const useStripeCheckout = () => {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [stripe, setStripe] = useState<Stripe | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    stripePromise.then((stripeInstance) => {
      setStripe(stripeInstance);
    }).catch((err) => {
      console.error('Failed to load Stripe:', err);
      setError('Failed to initialize payment system');
    });
  }, []);

  const createCheckoutSession = async (options: StripeCheckoutOptions) => {
    try {
      setIsLoading(true);
      setError(null);

      if (!stripe) {
        throw new Error('Stripe has not been initialized');
      }

      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('Authentication token not found');
      }

      // Call backend to create checkout session
      const response = await fetch('/api/subscriptions/stripe/checkout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          price_id: options.priceId,
          customer_email: options.customerEmail,
          success_url: options.successUrl || `${window.location.origin}/dashboard/subscriptions?success=true`,
          cancel_url: options.cancelUrl || `${window.location.origin}/dashboard/subscriptions/plans`,
          mode: options.mode || 'subscription',
          metadata: options.metadata || {},
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error?.message || 'Failed to create checkout session');
      }

      const data = await response.json();
      const sessionId = data.session_id;

      // Redirect to Stripe Checkout
      const { error: stripeError } = await stripe.redirectToCheckout({
        sessionId,
      });

      if (stripeError) {
        throw new Error(stripeError.message);
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Payment failed';
      setError(errorMessage);
      console.error('Stripe checkout error:', err);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const createSubscription = async (
    priceId: string,
    customerEmail: string,
    onSuccess?: (sessionId: string) => void,
    onError?: (error: Error) => void
  ) => {
    try {
      await createCheckoutSession({
        priceId,
        customerEmail,
        mode: 'subscription',
      });

      // Success callback will be handled by the redirect
      onSuccess?.(priceId);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Subscription creation failed');
      onError?.(error);
    }
  };

  const createPayment = async (
    priceId: string,
    customerEmail: string,
    amount?: number,
    onSuccess?: (sessionId: string) => void,
    onError?: (error: Error) => void
  ) => {
    try {
      await createCheckoutSession({
        priceId,
        customerEmail,
        mode: 'payment',
      });

      onSuccess?.(priceId);
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Payment creation failed');
      onError?.(error);
    }
  };

  const manageBilling = async () => {
    try {
      setIsLoading(true);
      setError(null);

      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('Authentication token not found');
      }

      // Call backend to create customer portal session
      const response = await fetch('/api/subscriptions/stripe/portal', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          return_url: `${window.location.origin}/dashboard/subscriptions`,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error?.message || 'Failed to open billing portal');
      }

      const data = await response.json();

      // Redirect to Stripe Customer Portal
      window.location.href = data.url;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to open billing portal';
      setError(errorMessage);
      console.error('Stripe portal error:', err);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    error,
    isReady: !!stripe,
    createCheckoutSession,
    createSubscription,
    createPayment,
    manageBilling,
  };
};
