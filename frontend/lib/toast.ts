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
export const updateToast = (toastId: string, options: any) => {
  return toast.success(options.render, { id: toastId });
};

// Helper to extract error message from API response
export const getErrorMessage = (error: any): string => {
  if (error?.response?.data?.message) {
    return error.response.data.message;
  }
  if (error?.response?.data?.error?.message) {
    return error.response.data.error.message;
  }
  if (error?.message) {
    return error.message;
  }
  return 'An unexpected error occurred';
};
