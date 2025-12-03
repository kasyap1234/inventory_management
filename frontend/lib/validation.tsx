// Form validation utility
export interface ValidationRule {
    required?: boolean;
    minLength?: number;
    maxLength?: number;
    pattern?: RegExp;
    min?: number;
    max?: number;
    custom?: (value: unknown) => string | null;
}

export interface ValidationRules {
    [key: string]: ValidationRule;
}

export interface ValidationErrors {
    [key: string]: string;
}

export function validateForm<T extends Record<string, unknown>>(
    data: T,
    rules: ValidationRules
): ValidationErrors {
    const errors: ValidationErrors = {};

    for (const [field, rule] of Object.entries(rules)) {
        const value = data[field];

        // Required validation
        if (rule.required && (value === null || value === undefined || value === '')) {
            errors[field] = `${formatFieldName(field)} is required`;
            continue;
        }

        // Skip other validations if value is empty and not required
        if (!value && !rule.required) continue;

        // String validations
        if (typeof value === 'string') {
            if (rule.minLength && value.length < rule.minLength) {
                errors[field] = `${formatFieldName(field)} must be at least ${rule.minLength} characters`;
            }
            if (rule.maxLength && value.length > rule.maxLength) {
                errors[field] = `${formatFieldName(field)} must not exceed ${rule.maxLength} characters`;
            }
            if (rule.pattern && !rule.pattern.test(value)) {
                errors[field] = `${formatFieldName(field)} format is invalid`;
            }
        }

        // Number validations
        if (typeof value === 'number') {
            if (rule.min !== undefined && value < rule.min) {
                errors[field] = `${formatFieldName(field)} must be at least ${rule.min}`;
            }
            if (rule.max !== undefined && value > rule.max) {
                errors[field] = `${formatFieldName(field)} must not exceed ${rule.max}`;
            }
        }

        // Custom validation
        if (rule.custom) {
            const customError = rule.custom(value);
            if (customError) {
                errors[field] = customError;
            }
        }
    }

    return errors;
}

function formatFieldName(field: string): string {
    return field
        .replace(/_/g, ' ')
        .replace(/([A-Z])/g, ' $1')
        .trim()
        .replace(/^\w/, (c) => c.toUpperCase());
}

// Common validation patterns
export const ValidationPatterns = {
    email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
    phone: /^[\d\s+()-]+$/,
    url: /^https?:\/\/.+/,
    uuid: /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    alphanumeric: /^[a-zA-Z0-9]+$/,
    numeric: /^\d+$/,
};

// Form field error display component
export function FieldError({ error }: { error?: string }) {
    if (!error) return null;
    return <p className="text-sm text-destructive mt-1">{error}</p>;
}
