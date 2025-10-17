'use client';

import * as React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

const ToastContext = React.createContext<{
  addToast: (message: string, type: ToastType) => void;
} | null>(null);

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<Toast[]>([]);

  const addToast = React.useCallback((message: string, type: ToastType) => {
    const id = Math.random().toString(36).substring(7);
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 5000);
  }, []);

  const removeToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <ToastContext.Provider value={{ addToast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-3">
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={cn(
              'flex items-center gap-3 px-5 py-4 rounded-xl shadow-elegant-lg min-w-[320px] animate-fade-in backdrop-blur-sm border-2',
              {
                'bg-gradient-to-r from-emerald-500 to-emerald-600 text-white border-emerald-400 shadow-colored': toast.type === 'success',
                'bg-gradient-to-r from-red-500 to-red-600 text-white border-red-400 shadow-colored': toast.type === 'error',
                'bg-gradient-to-r from-amber-500 to-amber-600 text-white border-amber-400 shadow-colored': toast.type === 'warning',
                'bg-gradient-to-r from-indigo-500 to-purple-600 text-white border-indigo-400 shadow-colored': toast.type === 'info',
              }
            )}
          >
            <p className="flex-1 text-sm font-semibold">{toast.message}</p>
            <button
              onClick={() => removeToast(toast.id)}
              className="text-white/90 hover:text-white hover:scale-110 transition-transform duration-200"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = React.useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return context;
}
