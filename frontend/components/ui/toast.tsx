'use client';

import * as React from 'react';
import { X, CheckCircle2, AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { cn } from '@/lib/utils';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
  id: string;
  message: string;
  type: ToastType;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
  duration?: number;
}

const ToastContext = React.createContext<{
  addToast: (message: string, type: ToastType, options?: Partial<Omit<Toast, 'id' | 'message' | 'type'>>) => void;
  removeToast: (id: string) => void;
} | null>(null);

const toastIcons = {
  success: CheckCircle2,
  error: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<Toast[]>([]);

  const addToast = React.useCallback((message: string, type: ToastType, options?: Partial<Omit<Toast, 'id' | 'message' | 'type'>>) => {
    const id = Math.random().toString(36).substring(7);
    const duration = options?.duration || 5000;
    const newToast = { id, message, type, ...options, duration };
    setToasts((prev) => [...prev, newToast]);
    
    if (duration > 0) {
      setTimeout(() => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
      }, duration);
    }
  }, []);

  const removeToast = React.useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ addToast, removeToast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-3 max-w-md" role="region" aria-label="Notifications">
        {toasts.map((toast) => {
          const Icon = toastIcons[toast.type];
          return (
            <div
              key={toast.id}
              role="alert"
              aria-live="polite"
              className={cn(
                'flex items-start gap-3 px-5 py-4 rounded-xl shadow-elegant-lg min-w-[320px] animate-fade-in backdrop-blur-sm border-2 transition-all duration-300 hover:shadow-2xl',
                {
                  'bg-gradient-to-r from-emerald-50 to-emerald-100 border-emerald-500/30 text-emerald-900': toast.type === 'success',
                  'bg-gradient-to-r from-red-50 to-red-100 border-red-500/30 text-red-900': toast.type === 'error',
                  'bg-gradient-to-r from-amber-50 to-amber-100 border-amber-500/30 text-amber-900': toast.type === 'warning',
                  'bg-gradient-to-r from-blue-50 to-blue-100 border-blue-500/30 text-blue-900': toast.type === 'info',
                }
              )}
            >
              <div className={cn(
                'flex-shrink-0 w-5 h-5 mt-0.5',
                {
                  'text-emerald-600': toast.type === 'success',
                  'text-red-600': toast.type === 'error',
                  'text-amber-600': toast.type === 'warning',
                  'text-blue-600': toast.type === 'info',
                }
              )}>
                <Icon className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold leading-tight">{toast.message}</p>
                {toast.description && (
                  <p className="text-xs mt-1 opacity-80">{toast.description}</p>
                )}
                {toast.action && (
                  <button
                    onClick={() => {
                      toast.action?.onClick();
                      removeToast(toast.id);
                    }}
                    className={cn(
                      'mt-2 text-xs font-semibold underline hover:no-underline transition-all',
                      {
                        'text-emerald-700 hover:text-emerald-800': toast.type === 'success',
                        'text-red-700 hover:text-red-800': toast.type === 'error',
                        'text-amber-700 hover:text-amber-800': toast.type === 'warning',
                        'text-blue-700 hover:text-blue-800': toast.type === 'info',
                      }
                    )}
                  >
                    {toast.action.label}
                  </button>
                )}
              </div>
              <button
                onClick={() => removeToast(toast.id)}
                className="flex-shrink-0 text-gray-500 hover:text-gray-700 hover:scale-110 transition-all duration-200"
                aria-label="Dismiss notification"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          );
        })}
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
