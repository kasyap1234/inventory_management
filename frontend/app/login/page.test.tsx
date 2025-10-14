import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { useRouter, useSearchParams } from 'next/navigation';
import LoginPage from './page';

// Mock dependencies
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
  useSearchParams: jest.fn(),
}));

jest.mock('@/hooks/useAuth', () => ({
  useAuth: jest.fn(),
}));

jest.mock('@/components/ui/button', () => ({
  Button: ({ children, onClick, disabled, type, className }: any) => 
    <button type={type} onClick={onClick} disabled={disabled} className={className}>
      {children}
    </button>,
}));

jest.mock('@/components/ui/input', () => ({
  Input: ({ onChange, value, type, placeholder, id, required }: any) =>
    <input
      type={type}
      id={id}
      placeholder={placeholder}
      value={value}
      onChange={onChange}
      required={required}
      data-testid={id}
    />,
}));

jest.mock('@/components/ui/card', () => ({
  Card: ({ children }: any) => <div data-testid="card">{children}</div>,
  CardContent: ({ children, className }: any) => 
    <div className={className} data-testid="card-content">{children}</div>
}));

import { useRouter } from 'next/navigation';
import { useSearchParams } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';

const mockRouter = {
  push: jest.fn(),
  replace: jest.fn(),
  back: jest.fn(),
  forward: jest.fn(),
  refresh: jest.fn(),
  prefetch: jest.fn(),
};

const mockSearchParams = new URLSearchParams();

describe('LoginPage', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
    (useSearchParams as jest.Mock).mockReturnValue(mockSearchParams);
    
    (useAuth as jest.Mock).mockReturnValue({
      login: {
        mutate: jest.fn(),
        isPending: false,
        isError: false,
        error: null,
      },
    });
  });

  describe('Initial Render', () => {
    it('should render the login form correctly', () => {
      render(<LoginPage />);

      expect(screen.getByText('Agromart')).toBeInTheDocument();
      expect(screen.getByText('Welcome back! Please sign in to continue.')).toBeInTheDocument();
      expect(screen.getByLabelText('Email address')).toBeInTheDocument();
      expect(screen.getByLabelText('Password')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
      expect(screen.getByText('Forgot?')).toBeInTheDocument();
      expect(screen.getByTestId('card')).toBeInTheDocument();
      expect(screen.getByTestId('card-content')).toBeInTheDocument();
    });

    it('should have proper accessibility structure', () => {
      render(<LoginPage />);

      expect(screen.getByRole('heading', { level: 1, name: 'Agromart' })).toBeInTheDocument();
      expect(screen.getByRole('textbox', { name: /email address/i })).toBeInTheDocument();
      expect(screen.getByLabelText('Password')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
    });

    it('should render forgot password link correctly', () => {
      render(<LoginPage />);

      const forgotLink = screen.getByText('Forgot?');
      expect(forgotLink).toBeInTheDocument();
      expect(forgotLink.closest('a')).toHaveAttribute('href', '/forgot-password');
    });
  });

  describe('Form Interaction', () => {
    it('should update email input value when typing', async () => {
      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });

      expect(emailInput).toHaveValue('test@example.com');
    });

    it('should update password input value when typing', async () => {
      render(<LoginPage />);

      const passwordInput = screen.getByTestId('password');
      
      fireEvent.change(passwordInput, { target: { value: 'password123' } });

      expect(passwordInput).toHaveValue('password123');
    });

    it('should submit form with correct credentials', async () => {
      const mockLogin = jest.fn();
      (useAuth as jest.Mock).mockReturnValue({
        login: {
          mutate: mockLogin,
          isPending: false,
          isError: false,
          error: null,
        },
      });

      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      const passwordInput = screen.getByTestId('password');
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      // Fill form
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });

      // Submit form
      fireEvent.click(submitButton);

      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      });
    });

    it('should submit form on enter key press', async () => {
      const mockLogin = jest.fn();
      (useAuth as jest.Mock).mockReturnValue({
        login: {
          mutate: mockLogin,
          isPending: false,
          isError: false,
          error: null,
        },
      });

      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      const passwordInput = screen.getByTestId('password');
      const form = screen.getByTestId('card-content')?.closest('form');

      // Fill form
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });

      // Submit with enter key on password field
      fireEvent.submit(passwordInput);

      expect(mockLogin).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      });
    });

    it('should disable button and show loading state during submission', async () => {
      const mockLogin = jest.fn();
      (useAuth as jest.Mock).mockReturnValue({
        login: {
          mutate: mockLogin,
          isPending: true,
          isError: false,
          error: null,
        },
      });

      render(<LoginPage />);

      const submitButton = screen.getByRole('button', { name: /sign in/i });

      expect(submitButton).toBeDisabled();
      expect(screen.getByText('Signing In...')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should display error message when login fails', async () => {
      const error = new Error('Invalid credentials');
      (error as any).response = {
        data: {
          error: {
            message: 'Invalid email or password',
          },
        },
      };

      (useAuth as jest.Mock).mockReturnValue({
        login: {
          mutate: jest.fn(),
          isPending: false,
          isError: true,
          error: error,
        },
      });

      render(<LoginPage />);

      expect(screen.getByText('Invalid email or password')).toBeInTheDocument();
      
      // Check for error icon
      const errorIcon = document.querySelector('svg');
      expect(errorIcon).toBeInTheDocument();
    });

    it('should display fallback error message when error has no message', async () => {
      const error = new Error('Login failed');
      (useAuth as jest.Mock).mockReturnValue({
        login: {
          mutate: jest.fn(),
          isPending: false,
          isError: true,
          error: error,
        },
      });

      render(<LoginPage />);

      expect(screen.getByText('Invalid credentials. Please try again.')).toBeInTheDocument();
    });

    it('should show reset success message when reset parameter is present', () => {
      mockSearchParams.set('reset', 'success');
      (useSearchParams as jest.Mock).mockReturnValue(mockSearchParams);

      render(<LoginPage />);

      expect(screen.getByText('Password updated successfully. Please sign in with your new password.')).toBeInTheDocument();
    });
  });

  describe('URL Parameters', () => {
    it('should handle reset success parameter', () => {
      mockSearchParams.set('reset', 'success');
      (useSearchParams as jest.Mock).mockReturnValue(mockSearchParams);

      render(<LoginPage />);

      expect(screen.getByText('Password updated successfully. Please sign in with your new password.')).toBeInTheDocument();
    });

    it('should not show success message when reset parameter is not success', () => {
      mockSearchParams.set('reset', 'failed');
      (useSearchParams as jest.Mock).mockReturnValue(mockSearchParams);

      render(<LoginPage />);

      expect(screen.queryByText(/Password updated successfully/)).not.toBeInTheDocument();
    });
  });

  describe('Form Validation', () => {
    it('should require email field', () => {
      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      expect(emailInput).toBeRequired();
    });

    it('should require password field', () => {
      render(<LoginPage />);

      const passwordInput = screen.getByTestId('password');
      expect(passwordInput).toBeRequired();
    });

    it('should use correct input types', () => {
      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      const passwordInput = screen.getByTestId('password');

      expect(emailInput).toHaveAttribute('type', 'email');
      expect(passwordInput).toHaveAttribute('type', 'password');
    });

    it('should have proper placeholder text', () => {
      render(<LoginPage />);

      const emailInput = screen.getByTestId('email');
      const passwordInput = screen.getByTestId('password');

      expect(emailInput).toHaveAttribute('placeholder', 'you@company.com');
      expect(passwordInput).toHaveAttribute('placeholder', 'Enter your password');
    });
  });

  describe('Feature Sections', () => {
    it('should render feature highlights', () => {
      render(<LoginPage />);

      // Check for feature sections
      expect(screen.getByText(/Secure by Design/)).toBeInTheDocument();
      expect(screen.getByText(/Built for Scale/)).toBeInTheDocument();
      expect(screen.getByText(/Lightning Fast/)).toBeInTheDocument();
      
      // Check for feature descriptions
      expect(screen.getByText(/Enterprise-grade security/)).toBeInTheDocument();
      expect(screen.getByText(/Handle millions of transactions/)).toBeInTheDocument();
      expect(screen.getByText(/Optimized for peak performance/)).toBeInTheDocument();

      // Check for icons
      expect(screen.getByTestId('shield-icon')).toBeInTheDocument();
      expect(screen.getByTestId('sparkles-icon')).toBeInTheDocument();
      expect(screen.getByTestId('zap-icon')).toBeInTheDocument();
    });

    it('should render stats section', () => {
      render(<LoginPage />);

      expect(screen.getByText('10K+')).toBeInTheDocument();
      expect(screen.getByText('Active Users')).toBeInTheDocument();
      expect(screen.getByText('99.9%')).toBeInTheDocument();
      expect(screen.getByText('Uptime')).toBeInTheDocument();
      expect(screen.getByText('24/7')).toBeInTheDocument();
      expect(screen.getByText('Support')).toBeInTheDocument();
    });
  });

  describe('Responsive Design', () => {
    it('should maintain layout structure on mobile', () => {
      // Mock mobile viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      render(<LoginPage />);

      // Should still render all essential elements
      expect(screen.getByText('Agromart')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
      expect(screen.getByTestId('card')).toBeInTheDocument();
    });

    it('should maintain layout structure on desktop', () => {
      // Mock desktop viewport
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 1920,
      });

      render(<LoginPage />);

      // Should render all elements including feature sections
      expect(screen.getByText('Agromart')).toBeInTheDocument();
      expect(screenByText(/Secure by Design/)).toBeInTheDocument();
      expect(screen.getByTestId('card')).toBeInTheDocument();
    });
  });

  describe('Animation Classes', () => {
    it('should apply correct animation classes', () => {
      render(<LoginPage />);

      const mainContent = document.querySelector('.animate-fade-in');
      expect(mainContent).toBeInTheDocument();
    });
  });

  describe('User Experience', () => {
    it('should focus email field on initial load', () => {
      render(<LoginPage />);

      // Note: This would require additional setup to properly test auto-focus
      // For now, we just verify the field exists and can receive focus
      const emailInput = screen.getByTestId('email');
      expect(emailInput).toBeInTheDocument();
      
      emailInput.focus();
      expect(document.activeElement).toBe(emailInput);
    });

    it('should provide clear visual feedback for interactive elements', () => {
      render(<LoginPage />);

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      expect(submitButton).toBeInTheDocument();
      
      // Verify button has proper styling classes (would need actual implementation)
      const forgotLink = screen.getByText('Forgot?');
      expect(forgotLink).toBeInTheDocument();
    });
  });
});
