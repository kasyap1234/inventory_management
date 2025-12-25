import { describe, it, expect } from 'vitest';
import { cn, formatCurrency, formatDate, formatDateTime } from './utils';

describe('cn utility function', () => {
    it('should merge class names correctly', () => {
        expect(cn('class1', 'class2')).toBe('class1 class2');
    });

    it('should handle conditional classes', () => {
        expect(cn('base', true && 'included', false && 'excluded')).toBe('base included');
    });

    it('should handle undefined and null values', () => {
        expect(cn('base', undefined, null, 'end')).toBe('base end');
    });

    it('should merge Tailwind classes correctly', () => {
        expect(cn('px-2 py-1', 'px-4')).toBe('py-1 px-4');
    });

    it('should handle arrays of classes', () => {
        expect(cn(['class1', 'class2'], 'class3')).toBe('class1 class2 class3');
    });

    it('should handle empty inputs', () => {
        expect(cn()).toBe('');
        expect(cn('')).toBe('');
    });

    it('should handle object syntax', () => {
        expect(cn({ 'text-red-500': true, 'text-blue-500': false })).toBe('text-red-500');
    });
});

describe('formatCurrency function', () => {
    it('should format positive amounts in INR', () => {
        const result = formatCurrency(1000);
        expect(result).toContain('1,000');
        expect(result).toContain('₹');
    });

    it('should format zero correctly', () => {
        const result = formatCurrency(0);
        expect(result).toContain('0');
    });

    it('should format decimal amounts', () => {
        const result = formatCurrency(1234.56);
        expect(result).toContain('1,234');
    });

    it('should handle different currencies', () => {
        const result = formatCurrency(1000, 'USD');
        expect(result).toContain('$');
    });

    it('should format large numbers with proper separators', () => {
        const result = formatCurrency(1000000);
        expect(result).toMatch(/10,00,000|1,000,000/); // Indian or standard format
    });

    it('should format negative amounts', () => {
        const result = formatCurrency(-500);
        // Negative amounts may be formatted with '-' or parentheses depending on locale
        const hasNegativeIndicator = result.includes('-') || result.includes('(');
        expect(hasNegativeIndicator).toBe(true);
    });
});

describe('formatDate function', () => {
    it('should format date string correctly', () => {
        const result = formatDate('2024-01-15');
        expect(result).toContain('15');
        expect(result).toContain('Jan');
        expect(result).toContain('2024');
    });

    it('should format Date object correctly', () => {
        const date = new Date('2024-06-20');
        const result = formatDate(date);
        expect(result).toContain('20');
        expect(result).toContain('Jun');
        expect(result).toContain('2024');
    });

    it('should handle end of year dates', () => {
        const result = formatDate('2024-12-31');
        expect(result).toContain('31');
        expect(result).toContain('Dec');
        expect(result).toContain('2024');
    });

    it('should handle beginning of year dates', () => {
        const result = formatDate('2024-01-01');
        expect(result).toContain('1');
        expect(result).toContain('Jan');
        expect(result).toContain('2024');
    });
});

describe('formatDateTime function', () => {
    it('should format date and time correctly', () => {
        const result = formatDateTime('2024-01-15T14:30:00');
        expect(result).toContain('15');
        expect(result).toContain('Jan');
        expect(result).toContain('2024');
    });

    it('should format Date object with time', () => {
        const date = new Date('2024-06-20T09:45:00');
        const result = formatDateTime(date);
        expect(result).toContain('20');
        expect(result).toContain('Jun');
        expect(result).toContain('2024');
    });

    it('should handle midnight', () => {
        const result = formatDateTime('2024-01-15T00:00:00');
        expect(result).toContain('15');
        expect(result).toContain('Jan');
    });

    it('should handle end of day', () => {
        const result = formatDateTime('2024-01-15T23:59:59');
        expect(result).toContain('15');
        expect(result).toContain('Jan');
    });
});
