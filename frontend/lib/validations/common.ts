import { z } from 'zod';

/**
 * UUID validation schema
 */
export const uuidSchema = z.string().uuid('Invalid ID format');

/**
 * Positive number validation
 */
export const positiveNumberSchema = z.number().positive('Must be a positive number');

/**
 * Non-negative number validation
 */
export const nonNegativeNumberSchema = z.number().nonnegative('Must be zero or greater');

/**
 * Positive integer validation
 */
export const positiveIntSchema = z.number().int('Must be a whole number').positive('Must be greater than zero');

/**
 * Email validation schema
 */
export const emailSchema = z.string().email('Invalid email address').toLowerCase().trim();

/**
 * Phone number validation (basic)
 */
export const phoneSchema = z.string().regex(/^[+]?[\d\s\-()]+$/, 'Invalid phone number format');

/**
 * URL validation schema
 */
export const urlSchema = z.string().url('Invalid URL format');

/**
 * Date string validation
 */
export const dateStringSchema = z.string().refine((val) => !isNaN(Date.parse(val)), {
  message: 'Invalid date format',
});

/**
 * Pagination parameters validation
 */
export const paginationSchema = z.object({
  limit: z.number().int().min(1).max(100).default(20),
  offset: z.number().int().min(0).default(0),
});

export type PaginationParams = z.infer<typeof paginationSchema>;

/**
 * Search query validation
 */
export const searchQuerySchema = z.object({
  q: z.string().min(1, 'Search query is required').max(100, 'Search query is too long').trim(),
});

export type SearchQuery = z.infer<typeof searchQuerySchema>;

/**
 * File upload validation
 */
export const fileUploadSchema = z.object({
  file: z.instanceof(File)
    .refine((file) => file.size <= 10 * 1024 * 1024, 'File size must be less than 10MB')
    .refine(
      (file) => ['image/jpeg', 'image/png', 'image/webp', 'image/gif'].includes(file.type),
      'File must be an image (JPEG, PNG, WebP, or GIF)'
    ),
});

export type FileUpload = z.infer<typeof fileUploadSchema>;

/**
 * Price/Amount validation (supports decimals)
 */
export const priceSchema = z.number()
  .nonnegative('Price must be zero or greater')
  .max(999999999.99, 'Price is too large')
  .refine((val) => {
    // Check for at most 2 decimal places
    const decimalPlaces = (val.toString().split('.')[1] || '').length;
    return decimalPlaces <= 2;
  }, 'Price can have at most 2 decimal places');

/**
 * Quantity validation
 */
export const quantitySchema = z.number()
  .int('Quantity must be a whole number')
  .nonnegative('Quantity must be zero or greater')
  .max(999999, 'Quantity is too large');

/**
 * Percentage validation (0-100)
 */
export const percentageSchema = z.number()
  .min(0, 'Percentage must be between 0 and 100')
  .max(100, 'Percentage must be between 0 and 100');

/**
 * Status enum validation
 */
export const statusSchema = z.enum(['active', 'inactive', 'pending', 'deleted']);

export type Status = z.infer<typeof statusSchema>;

/**
 * Text area validation (with max length)
 */
export const textAreaSchema = z.string()
  .max(1000, 'Text is too long (maximum 1000 characters)')
  .trim();

/**
 * Optional string that trims and converts empty to undefined
 */
export const optionalStringSchema = z.string()
  .trim()
  .transform((val) => val === '' ? undefined : val)
  .optional();
