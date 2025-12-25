import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { PasswordStrengthMeter } from '@/components/password-strength-meter';

// Mock the password library
vi.mock('@/lib/password', () => ({
  evaluatePasswordStrength: vi.fn((password: string) => {
    if (password.length < 4) {
      return {
        score: 1,
        label: 'Weak',
        isAcceptable: false,
        suggestions: ['Add more characters'],
      };
    }
    if (password.length < 8) {
      return {
        score: 2,
        label: 'Fair',
        isAcceptable: false,
        suggestions: ['Add uppercase letters'],
      };
    }
    if (password.length < 12) {
      return {
        score: 3,
        label: 'Strong',
        isAcceptable: true,
        suggestions: [],
      };
    }
    return {
      score: 4,
      label: 'Very Strong',
      isAcceptable: true,
      suggestions: [],
    };
  }),
}));

describe('PasswordStrengthMeter', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns null when password is empty', () => {
    const { container } = render(<PasswordStrengthMeter password="" />);
    expect(container.firstChild).toBeNull();
  });

  it('displays strength label', () => {
    render(<PasswordStrengthMeter password="abc" />);
    expect(screen.getByText('Weak')).toBeInTheDocument();
  });

  it('displays Fair for medium passwords', () => {
    render(<PasswordStrengthMeter password="abcd123" />);
    expect(screen.getByText('Fair')).toBeInTheDocument();
  });

  it('displays Strong for good passwords', () => {
    render(<PasswordStrengthMeter password="MyP@ssw0rd" />);
    expect(screen.getByText('Strong')).toBeInTheDocument();
  });

  it('displays Very Strong for excellent passwords', () => {
    render(<PasswordStrengthMeter password="MyV3ryStr0ng!Password" />);
    expect(screen.getByText('Very Strong')).toBeInTheDocument();
  });

  it('shows suggestions for weak passwords', () => {
    render(<PasswordStrengthMeter password="abc" />);
    expect(screen.getByText(/Suggestions:/)).toBeInTheDocument();
    expect(screen.getByText(/Add more characters/)).toBeInTheDocument();
  });

  it('hides suggestions for strong passwords', () => {
    render(<PasswordStrengthMeter password="MyP@ssw0rd" />);
    expect(screen.queryByText(/Suggestions:/)).not.toBeInTheDocument();
  });

  it('renders strength bars', () => {
    const { container } = render(<PasswordStrengthMeter password="test1234" />);
    const bars = container.querySelectorAll('.flex-1');
    expect(bars.length).toBeGreaterThan(0);
  });
});
