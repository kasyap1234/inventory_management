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
                'flex items-start gap-3 px-5 py-4 rounded-xl shadow-lg min-w-[320px] border transition-all duration-300',
                {
                  'bg-background border-border text-foreground': true, // Default base styles
                  'border-l-4 border-l-green-500': toast.type === 'success',
                  'border-l-4 border-l-destructive': toast.type === 'error',
                  'border-l-4 border-l-yellow-500': toast.type === 'warning',
                  'border-l-4 border-l-blue-500': toast.type === 'info',
                }
              )}
            >
              <div className={cn(
                'flex-shrink-0 w-5 h-5 mt-0.5',
                {
                  'text-green-500': toast.type === 'success',
                  'text-destructive': toast.type === 'error',
                  'text-yellow-500': toast.type === 'warning',
                  'text-blue-500': toast.type === 'info',
                }
              )}>
                <Icon className="w-5 h-5" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-semibold leading-tight">{toast.message}</p>
                {toast.description && (
                  <p className="text-xs mt-1 text-muted-foreground">{toast.description}</p>
                )}
                {toast.action && (
                  <button
                    onClick={() => {
                      toast.action?.onClick();
                      removeToast(toast.id);
                    }}
                    className="mt-2 text-xs font-semibold underline hover:no-underline transition-all text-primary"
                  >
                    {toast.action.label}
                  </button>
                )}
              </div>
              <button
                onClick={() => removeToast(toast.id)}
                className="flex-shrink-0 text-muted-foreground hover:text-foreground transition-colors"
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
