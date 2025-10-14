import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { ThemeToggle } from './theme-toggle';

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};

  return {
    getItem: jest.fn((key: string) => store[key] || null),
    setItem: jest.fn((key: string, value: string) => {
      store[key] = value.toString();
    }),
    removeItem: jest.fn((key: string) => {
      delete store[key];
    }),
    clear: jest.fn(() => {
      store = {};
    }),
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});

describe('ThemeToggle', () => {
  beforeEach(() => {
    localStorageMock.clear();
    document.documentElement.className = '';
    jest.clearAllMocks();
  });

  afterEach(() => {
    document.documentElement.className = '';
  });

  describe('Initial Theme Setup', () => {
    it('should default to light theme when no preference is saved and system prefers light', () => {
      // Mock system prefers light
      window.matchMedia = jest.fn().mockImplementation(query => ({
        matches: query === '(prefers-color-scheme: light)',
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      }));

      render(<ThemeToggle />);

      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('theme');
    });

    it('should default to dark theme when no preference is saved and system prefers dark', () => {
      // Mock system prefers dark
      window.matchMedia = jest.fn().mockImplementation(query => ({
        matches: query === '(prefers-color-scheme: dark)',
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      }));

      render(<ThemeToggle />);

      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('theme');
    });

    it('should use saved theme preference when available', () => {
      localStorageMock.setItem('theme', 'dark');

      render(<ThemeToggle />);

      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('theme');
    });

    it('should use saved light theme preference when available', () => {
      localStorageMock.setItem('theme', 'light');

      render(<ThemeToggle />);

      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('theme');
    });
  });

  describe('Theme Toggle Functionality', () => {
    it('should toggle from light to dark theme on click', () => {
      // Start with light theme
      window.matchMedia = jest.fn().mockImplementation(query => ({
        matches: query === '(prefers-color-scheme: light)',
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      }));

      render(<ThemeToggle />);

      // Verify initial state is light
      expect(document.documentElement.classList.contains('dark')).toBe(false);

      // Click to toggle
      const button = screen.getByRole('button', { name: /toggle theme/i });
      fireEvent.click(button);

      // Should now be dark
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
    });

    it('should toggle from dark to light theme on click', () => {
      // Start with dark theme
      localStorageMock.setItem('theme', 'dark');
      window.matchMedia = jest.fn().mockImplementation(query => ({
        matches: false, // System prefers light but saved is dark
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      }));

      render(<ThemeToggle />);

      // Verify initial state is dark
      expect(document.documentElement.classList.contains('dark')).toBe(true);

      // Click to toggle
      const button = screen.getByRole('button', { name: /toggle theme/i });
      fireEvent.click(button);

      // Should now be light
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
    });

    it('should toggle multiple times correctly', () => {
      window.matchMedia = jest.fn().mockImplementation(query => ({
        matches: query === '(prefers-color-scheme: light)',
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
      }));

      render(<ThemeToggle />);

      const button = screen.getByRole('button', { name: /toggle theme/i });

      // Initial: light
      expect(document.documentElement.classList.contains('dark')).toBe(false);

      // First click: dark
      fireEvent.click(button);
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorageMock.setItem).toHaveBeenLastCalledWith('theme', 'dark');

      // Second click: light
      fireEvent.click(button);
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(localStorageMock.setItem).toHaveBeenLastCalledWith('theme', 'light');

      // Third click: dark
      fireEvent.click(button);
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      expect(localStorageMock.setItem).toHaveBeenLastCalledWith('theme', 'dark');
    });
  });

  describe('Accessibility', () => {
    it('should have proper aria-label', () => {
      render(<ThemeToggle />);

      const button = screen.getByRole('button', { name: /toggle theme/i });
      expect(button).toHaveAttribute('aria-label', 'Toggle theme');
    });

    it('should have screen reader text', () => {
      render(<ThemeToggle />);

      const srText = screen.getByText('Toggle theme');
      expect(srText).toHaveClass('sr-only');
    });
  });

  describe('Component Structure', () => {
    it('should render button with Sun and Moon icons', () => {
      render(<ThemeToggle />);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
      
      // Both icons should be present (visible/hidden controlled by CSS classes)
      const sunIcon = button.querySelector('.lucide-sun');
      const moonIcon = button.querySelector('.lucide-moon');
      
      expect(sunIcon).toBeInTheDocument();
      expect(moonIcon).toBeInTheDocument();
    });

    it('should have correct CSS classes', () => {
      render(<ThemeToggle />);

      const button = screen.getByRole('button');
      expect(button).toHaveClass(
        'rounded-full',
        'hover:bg-accent',
        'transition-colors'
      );
    });
  });

  describe('LocalStorage Integration', () => {
    it('should save theme preference to localStorage', () => {
      render(<ThemeToggle />);

      const button = screen.getByRole('button', { name: /toggle theme/i });
      fireEvent.click(button);

      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'dark');
    });

    it('should save light theme preference to localStorage', () => {
      localStorageMock.setItem('theme', 'dark');
      render(<ThemeToggle />);

      const button = screen.getByRole('button', { name: /toggle theme/i });
      fireEvent.click(button);

      expect(localStorageMock.setItem).toHaveBeenCalledWith('theme', 'light');
    });
  });

  describe('Edge Cases', () => {
    it('should handle corrupted localStorage data gracefully', () => {
      localStorageMock.setItem('theme', 'invalid');

      render(<ThemeToggle />);

      // Should fall back to system preference
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });

    it('should handle localStorage being unavailable', () => {
      // Mock localStorage to throw error
      const originalGetItem = localStorageMock.getItem;
      localStorageMock.getItem = jest.fn(() => {
        throw new Error('localStorage unavailable');
      });

      const consoleSpy = jest.spyOn(console, 'error').mockImplementation();

      render(<ThemeToggle />);

      // Should not crash and should default to light theme
      expect(document.documentElement.classList.contains('dark')).toBe(false);

      consoleSpy.mockRestore();
      localStorageMock.getItem = originalGetItem;
    });
  });

  describe('CSS Transitions', () => {
    it('should apply correct CSS classes during animation', () => {
      render(<ThemeToggle />);

      const button = screen.getByRole('button');
      const sunIcon = button.querySelector('.lucide-sun');
      const moonIcon = button.querySelector('.lucide-moon');

      // Initial state (light theme)
      expect(sunIcon).toHaveClass(
        'rotate-0',
        'scale-100',
        'transition-all'
      );
      expect(sunIcon).not.toHaveClass('dark:-rotate-90', 'dark:scale-0');
      
      expect(moonIcon).toHaveClass(
        'absolute',
        'rotate-90',
        'scale-0',
        'transition-all'
      );
      expect(moonIcon).not.toHaveClass('dark:rotate-0', 'dark:scale-100');
    });
  });
});
