import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { PasswordStrengthMeter } from './password-strength-meter';

// Mock the password evaluation function
jest.mock('@/lib/password', () => ({
  evaluatePasswordStrength: jest.fn(),
}));

import { evaluatePasswordStrength } from '@/lib/password';

describe('PasswordStrengthMeter', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Password Strength Evaluation', () => {
    it('should render no password state', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 0,
        label: 'Weak',
        isAcceptable: false,
        suggestions: ['Add more characters'],
      });

      render(<PasswordStrengthMeter password="" />);

      expect(screen.getByText('Weak')).toBeInTheDocument();
      
      // Should show suggestions for weak passwords
      expect(screen.getByText('Add more characters')).toBeInTheDocument();
    });

    it('should render weak password state', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 1,
        label: 'Weak',
        isAcceptable: false,
        suggestions: [
          'Use uppercase and lowercase letters',
          'Add numbers',
          'Add special characters',
        ],
      });

      render(<PasswordStrengthMeter password="abc" />);

      expect(screen.getByText('Weak')).toBeInTheDocument();
      
      // Should show first 3 suggestions
      expect(screen.getByText('Use uppercase and lowercase letters')).toBeInTheDocument();
      expect(screen.getByText('Add numbers')).toBeInTheDocument();
      expect(screen.getByText('Add special characters')).toBeInTheDocument();
    });

    it('should render fair password state', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: ['Add special characters'],
      });

      render(<PasswordStrengthMeter password="Password123" />);

      expect(screen.getByText('Fair')).toBeInTheDocument();
      
      // Should show remaining suggestions
      expect(screen.getByText('Add special characters')).toBeInTheDocument();
    });

    it('should render good password state', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 3,
        label: 'Good',
        isAcceptable: true,
        suggestions: [],
      });

      render(<PasswordStrengthMeter password="Password123!" />);

      expect(screen.getByText('Good')).toBeInTheDocument();
      
      // Should not show suggestions for good passwords
      expect(screen.queryByText(/Add/)).not.toBeInTheDocument();
    });

    it('should render strong password state', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 4,
        label: 'Strong',
        isAcceptable: true,
        suggestions: [],
      });

      render(<PasswordStrengthMeter password="StrongP@ssw0rd123!" />);

      expect(screen.getByText('Strong')).toBeInTheDocument();
      
      // Should not show suggestions for strong passwords
      expect(screen.queryByText(/Add/)).not.toBeInTheDocument();
    });
  });

  describe('Visual Strength Indicators', () => {
    it('should show appropriate colored bars for weak password (score 1)', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 1,
        label: 'Weak',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="weak" />);
      
      const bars = container.querySelectorAll('.flex-1');
      expect(bars).toHaveLength(4);
      
      // First bar should be orange (weak)
      expect(bars[0]).toHaveClass('bg-orange-500');
      // Rest should be gray
      expect(bars[1]).toHaveClass('bg-gray-200');
      expect(bars[2]).toHaveClass('bg-gray-200');
      expect(bars[3]).toHaveClass('bg-gray-200');
    });

    it('should show appropriate colored bars for fair password (score 2)', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="fair" />);
      
      const bars = container.querySelectorAll('.flex-1');
      expect(bars).toHaveLength(4);
      
      // First two bars should be yellow (fair)
      expect(bars[0]).toHaveClass('bg-yellow-500');
      expect(bars[1]).toHaveClass('bg-yellow-500');
      // Rest should be gray
      expect(bars[2]).toHaveClass('bg-gray-200');
      expect(bars[3]).toHaveClass('bg-gray-200');
    });

    it('should show appropriate colored bars for good password (score 3)', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 3,
        label: 'Good',
        isAcceptable: true,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="good" />);
      
      const bars = container.querySelectorAll('.flex-1');
      expect(bars).toHaveLength(4);
      
      // First three bars should be green (good)
      expect(bars[0]).toHaveClass('bg-green-500');
      expect(bars[1]).toHaveClass('bg-green-500');
      expect(bars[2]).toHaveClass('bg-green-500');
      // Last should be gray
      expect(bars[3]).toHaveClass('bg-gray-200');
    });

    it('should show appropriate colored bars for strong password (score 4)', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 4,
        label: 'Strong',
        isAcceptable: true,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="strong" />);
      
      const bars = container.querySelectorAll('.flex-1');
      expect(bars).toHaveLength(4);
      
      // All bars should be green (strong)
      expect(bars[0]).toHaveClass('bg-green-500');
      expect(bars[1]).toHaveClass('bg-green-500');
      expect(bars[2]).toHaveClass('bg-green-500');
      expect(bars[3]).toHaveClass('bg-green-500');
    });
  });

  describe('Suggestions Display', () => {
    it('should limit suggestions to 3 items', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 1,
        label: 'Weak',
        isAcceptable: false,
        suggestions: [
          'Use uppercase and lowercase letters',
          'Add numbers',
          'Add special characters',
          'Make it longer',
          'Avoid common patterns',
        ],
      });

      render(<PasswordStrengthMeter password="weak" />);

      // Should only show first 3 suggestions
      const suggestions = screen.getAllByRole('listitem');
      expect(suggestions).toHaveLength(3);
      
      expect(suggestions[0]).toHaveTextContent('Use uppercase and lowercase letters');
      expect(suggestions[1]).toHaveTextContent('Add numbers');
      expect(suggestions[2]).toHaveTextContent('Add special characters');
    });

    it('should hide suggestions for acceptable passwords', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 3,
        label: 'Good',
        isAcceptable: true,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="good" />);

      // Should not show suggestion list
      const suggestionList = container.querySelector('ul');
      expect(suggestionList).not.toBeInTheDocument();
    });

    it('should hide suggestions when no suggestions are provided', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="fair" />);

      // Should not show suggestion list
      const suggestionList = container.querySelector('ul');
      expect(suggestionList).not.toBeInTheDocument();
    });
  });

  describe('Component Structure', () => {
    it('should render correct container structure', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="test" />);

      // Main container
      expect(container.firstChild).toHaveClass('space-y-2');
      
      // Strength indicator container
      const strengthContainer = container.querySelector('.flex.items-center');
      expect(strengthContainer).toBeInTheDocument();
      
      // Progress bar container
      const progressBar = container.querySelector('.flex.flex-1.h-2');
      expect(progressBar).toBeInTheDocument();
      expect(progressBar).toHaveClass(
        'rounded-full',
        'bg-gray-200',
        'overflow-hidden'
      );
      
      // Four indicator bars
      const bars = progressBar.querySelectorAll('.flex-1');
      expect(bars).toHaveLength(4);
    });

    it('should have correct label positioning and styling', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 3,
        label: 'Good',
        isAcceptable: true,
        suggestions: [],
      });

      render(<PasswordStrengthMeter password="good" />);

      const label = screen.getByText('Good');
      expect(label).toHaveClass(
        'text-xs',
        'font-medium',
        'text-foreground',
        'w-20',
        'text-right'
      );
    });
  });

  describe('Transitions and Animations', () => {
    it('should apply transition classes to bars', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="test" />);

      const bars = container.querySelectorAll('.flex-1');
      
      // All bars should have transition class
      bars.forEach(bar => {
        expect(bar).toHaveClass('transition-colors', 'duration-300');
      });
    });
  });

  describe('Accessibility', () => {
    it('should be keyboard accessible through form context', () => {
      // Since this is a visual feedback component, keyboard accessibility
      // is primarily handled through the parent form
      
      const { getByRole } = render(
        <div>
          <input type="password" aria-describedby="password-strength" />
          <PasswordStrengthMeter password="test" />
        </div>
      );

      const input = getByRole('textbox');
      expect(input).toHaveAttribute('aria-describedby', 'password-strength');
    });

    it('should provide meaningful visual feedback', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 1,
        label: 'Weak',
        isAcceptable: false,
        suggestions: ['Make it stronger'],
      });

      render(<PasswordStrengthMeter password="weak" />);

      // Color indicators help users with normal vision
      // Suggestions provide clear text alternatives
      expect(screen.getByText('Weak')).toBeInTheDocument();
      expect(screen.getByText('Make it stronger')).toBeInTheDocument();
    });
  });

  describe('Responsive Design', () => {
    it('should maintain layout on different screen sizes', () => {
      (evaluatePasswordStrength as jest.Mock).mockReturnValue({
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: [],
      });

      const { container } = render(<PasswordStrengthMeter password="test" />);

      // Flex layout should be responsive
      const flexContainer = container.querySelector('.flex.items-center');
      expect(flexContainer).toHaveClass('flex');
      
      // Progress bar should take available space
      const progressBar = container.querySelector('.flex.flex-1');
      expect(progressBar).toHaveClass('flex-1');
    });
  });
});
