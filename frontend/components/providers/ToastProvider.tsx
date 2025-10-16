'use client';

import { ToastProvider as InternalToastProvider } from '@/components/ui/toast';
import { Toaster } from 'react-hot-toast';

type Props = {
  children: React.ReactNode;
};

export default function ToastProvider({ children }: Props) {
  return (
    <InternalToastProvider>
      {children}
      <Toaster
        position="top-right"
        reverseOrder={false}
        gutter={8}
        containerClassName=""
        containerStyle={{}}
        toastOptions={{
          className: '',
          duration: 4000,
          style: {
            background: '#363636',
            color: '#fff',
          },
          success: {
            duration: 3000,
            iconTheme: {
              primary: '#10b981',
              secondary: '#fff',
            },
            style: {
              background: '#ecfdf5',
              color: '#065f46',
              border: '1px solid #10b981',
            },
          },
          error: {
            duration: 5000,
            iconTheme: {
              primary: '#ef4444',
              secondary: '#fff',
            },
            style: {
              background: '#fef2f2',
              color: '#991b1b',
              border: '1px solid #ef4444',
            },
          },
          loading: {
            iconTheme: {
              primary: '#3b82f6',
              secondary: '#fff',
            },
            style: {
              background: '#eff6ff',
              color: '#1e40af',
              border: '1px solid #3b82f6',
            },
          },
        }}
      />
    </InternalToastProvider>
  );
}
