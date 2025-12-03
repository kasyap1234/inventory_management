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
export const updateToast = (toastId: string, options: { render: string }) => {
  return toast.success(options.render, { id: toastId });
};

interface APIError {
  response?: {
    data?: {
      message?: string;
      error?: {
        message?: string;
      };
    };
  };
  message?: string;
}

// Helper to extract error message from API response
export const getErrorMessage = (error: unknown): string => {
  const apiError = error as APIError;

  if (apiError?.response?.data?.message) {
    return apiError.response.data.message;
  }
  if (apiError?.response?.data?.error?.message) {
    return apiError.response.data.error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'An unexpected error occurred';
};
