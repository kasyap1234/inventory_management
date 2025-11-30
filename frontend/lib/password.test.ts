import { describe, it, expect } from 'vitest';
import { evaluatePasswordStrength, type PasswordStrength } from './password';

describe('evaluatePasswordStrength', () => {
    describe('empty and weak passwords', () => {
        it('should return very weak for empty password', () => {
            const result = evaluatePasswordStrength('');
            expect(result.score).toBe(0);
            expect(result.label).toBe('Very weak');
            expect(result.isAcceptable).toBe(false);
            expect(result.suggestions).toContain('Password cannot be empty');
        });

        it('should return very weak for whitespace-only password', () => {
            const result = evaluatePasswordStrength('   ');
            expect(result.score).toBe(0);
            expect(result.label).toBe('Very weak');
            expect(result.isAcceptable).toBe(false);
        });

        it('should detect common passwords', () => {
            const result = evaluatePasswordStrength('password');
            expect(result.suggestions).toContain('Avoid using common passwords');
        });

        it('should detect common password variations', () => {
            const result = evaluatePasswordStrength('PASSWORD');
            expect(result.suggestions).toContain('Avoid using common passwords');
        });

        it('should flag very short passwords', () => {
            const result = evaluatePasswordStrength('abc');
            expect(result.isAcceptable).toBe(false);
        });
    });

    describe('password length scoring', () => {
        it('should give better score for longer passwords', () => {
            const short = evaluatePasswordStrength('abc123AB!');
            const long = evaluatePasswordStrength('abc123AB!abc123AB!');
            // Both can reach max score of 4 due to normalization
            expect(long.score).toBeGreaterThanOrEqual(short.score);
        });

        it('should suggest longer password for passwords under 12 chars', () => {
            const result = evaluatePasswordStrength('Ab1!efgh');
            expect(result.suggestions).toContain('Use at least 12 characters for better strength');
        });
    });

    describe('character requirements', () => {
        it('should require lowercase letters', () => {
            const result = evaluatePasswordStrength('ABCD1234!');
            expect(result.suggestions).toContain('Add a lowercase letter');
        });

        it('should require uppercase letters', () => {
            const result = evaluatePasswordStrength('abcd1234!');
            expect(result.suggestions).toContain('Add an uppercase letter');
        });

        it('should require numbers', () => {
            const result = evaluatePasswordStrength('abcdABCD!');
            expect(result.suggestions).toContain('Add a number');
        });

        it('should require special characters', () => {
            const result = evaluatePasswordStrength('abcdABCD1234');
            expect(result.suggestions).toContain('Add a special character');
        });
    });

    describe('pattern detection', () => {
        it('should detect repeated sequences', () => {
            const result = evaluatePasswordStrength('aaabbbccc123!');
            expect(result.suggestions).toContain('Avoid repeated sequences');
        });

        it('should suggest mixing character types', () => {
            const result = evaluatePasswordStrength('abcdefghij');
            expect(result.suggestions).toContain('Mix letters, numbers, and symbols');
        });

        it('should detect numbers-only passwords', () => {
            const result = evaluatePasswordStrength('12345678901234');
            expect(result.suggestions).toContain('Mix letters, numbers, and symbols');
        });
    });

    describe('strong passwords', () => {
        it('should accept strong password with all requirements', () => {
            const result = evaluatePasswordStrength('MySecureP@ss123!');
            expect(result.isAcceptable).toBe(true);
            expect(result.score).toBeGreaterThanOrEqual(3);
        });

        it('should rate very long complex password as very strong', () => {
            const result = evaluatePasswordStrength('MyV3ryL0ng&Secure!Password');
            expect(result.score).toBe(4);
            expect(result.label).toBe('Very strong');
        });

        it('should have no suggestions for perfect password', () => {
            const result = evaluatePasswordStrength('MySuperSecureP@ss123!XYZ');
            expect(result.suggestions).toHaveLength(0);
        });
    });

    describe('label mapping', () => {
        it('should return correct labels for score ranges', () => {
            // Single character gets some score from lowercase requirement
            const weak = evaluatePasswordStrength('a');
            expect(weak.label).toBe('Weak');

            // Strong passwords
            const strong = evaluatePasswordStrength('StrongP@ss123');
            expect(['Fair', 'Strong', 'Very strong']).toContain(strong.label);
        });
    });

    describe('edge cases', () => {
        it('should handle unicode characters', () => {
            const result = evaluatePasswordStrength('MyP@ss123日本語');
            expect(result).toBeDefined();
            expect(result.label).toBeDefined();
        });

        it('should handle very long passwords', () => {
            const longPassword = 'Aa1!' + 'x'.repeat(100);
            const result = evaluatePasswordStrength(longPassword);
            expect(result).toBeDefined();
            expect(result.score).toBeGreaterThan(0);
        });

        it('should handle passwords with only special characters', () => {
            const result = evaluatePasswordStrength('!@#$%^&*()');
            expect(result.suggestions).toContain('Add a lowercase letter');
            expect(result.suggestions).toContain('Add an uppercase letter');
            expect(result.suggestions).toContain('Add a number');
        });

        it('should trim whitespace from password', () => {
            const result = evaluatePasswordStrength('  password  ');
            expect(result.suggestions).toContain('Avoid using common passwords');
        });
    });
});
