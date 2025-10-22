import { cn } from '@/lib/utils';
import { Loader2 } from 'lucide-react';

interface SpinnerProps {
  className?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  variant?: 'default' | 'primary' | 'white';
  label?: string;
}

const sizeMap = {
  xs: 'h-3 w-3',
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
  xl: 'h-12 w-12',
};

const variantMap = {
  default: 'text-gray-600',
  primary: 'text-blue-600',
  white: 'text-white',
};

export function Spinner({ className, size = 'md', variant = 'default', label }: SpinnerProps) {
  return (
    <div className="flex items-center gap-2" role="status" aria-label={label || 'Loading'}>
      <Loader2
        className={cn(
          'animate-spin',
          sizeMap[size],
          variantMap[variant],
          className
        )}
        aria-hidden="true"
      />
      {label && (
        <span className={cn('text-sm font-medium', variantMap[variant])}>{label}</span>
      )}
      <span className="sr-only">{label || 'Loading...'}</span>
    </div>
  );
}

interface SpinnerOverlayProps {
  message?: string;
  variant?: 'light' | 'dark';
}

export function SpinnerOverlay({ message, variant = 'light' }: SpinnerOverlayProps) {
  return (
    <div 
      className={cn(
        'fixed inset-0 z-50 flex flex-col items-center justify-center backdrop-blur-sm transition-all duration-300',
        variant === 'light' ? 'bg-white/80' : 'bg-black/60'
      )}
      role="alert"
      aria-busy="true"
      aria-label={message || 'Loading'}
    >
      <div className="flex flex-col items-center gap-4 p-8 bg-white rounded-2xl shadow-2xl animate-fade-in">
        <Spinner size="xl" variant="primary" />
        {message && (
          <p className="text-sm font-medium text-gray-700 text-center max-w-xs">
            {message}
          </p>
        )}
      </div>
    </div>
  );
}

export function LoadingScreen({ message = "Loading..." }: { message?: string }) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gradient-to-br from-white to-gray-50">
      <Spinner size="lg" />
      <p className="mt-4 text-gray-600 font-medium">{message}</p>
    </div>
  )
}
