import toast from 'react-hot-toast';

// Success toast
export const showSuccess = (message: string) => {
  return toast.success(message);
};

// Error toast
export const showError = (message: string) => {
  return toast.error(message);
};

// Loading toast
export const showLoading = (message: string) => {
  return toast.loading(message);
};

// Promise toast - automatically handles loading, success, and error states
export const showPromise = <T,>(
  promise: Promise<T>,
  messages: {
    loading: string;
    success: string;
    error: string;
  }
) => {
  return toast.promise(promise, messages);
};

// Custom toast
export const showToast = (message: string) => {
  return toast(message);
};

// Dismiss a specific toast
export const dismissToast = (toastId: string) => {
  return toast.dismiss(toastId);
};

// Dismiss all toasts
export const dismissAllToasts = () => {
  return toast.dismiss();
};

// Update an existing toast
export const updateToast = (toastId: string, message: string) => {
  return toast.success(message, { id: toastId });
};

// Helper to extract error message from API response
export const getErrorMessage = (error: unknown): string => {
  // Type-safe error extraction
  if (error && typeof error === 'object') {
    const err = error as Record<string, unknown>;
    
    // Check for nested error structures
    if (err.response && typeof err.response === 'object') {
      const response = err.response as Record<string, unknown>;
      if (response.data && typeof response.data === 'object') {
        const data = response.data as Record<string, unknown>;
        
        if (typeof data.message === 'string') {
          return data.message;
        }
        
        if (data.error && typeof data.error === 'object') {
          const errorObj = data.error as Record<string, unknown>;
          if (typeof errorObj.message === 'string') {
            return errorObj.message;
          }
        }
      }
    }
    
    // Check for direct message property
    if (typeof err.message === 'string') {
      return err.message;
    }
  }
  
  return 'An unexpected error occurred';
};
