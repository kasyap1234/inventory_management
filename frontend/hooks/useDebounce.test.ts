import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDebounce } from './useDebounce';

interface TestProps {
    value: string;
    delay?: number;
}

describe('useDebounce', () => {
    beforeEach(() => {
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('should return initial value immediately', () => {
        const { result } = renderHook(() => useDebounce('initial', 500));
        expect(result.current).toBe('initial');
    });

    it('should debounce value changes', async () => {
        const { result, rerender } = renderHook(
            ({ value, delay }: TestProps) => useDebounce(value, delay),
            { initialProps: { value: 'initial', delay: 500 } }
        );

        expect(result.current).toBe('initial');

        // Update the value
        rerender({ value: 'updated', delay: 500 });

        // Value should still be initial immediately
        expect(result.current).toBe('initial');

        // Fast forward time
        act(() => {
            vi.advanceTimersByTime(500);
        });

        // Now it should be updated
        expect(result.current).toBe('updated');
    });

    it('should cancel previous timeout on rapid updates', () => {
        const { result, rerender } = renderHook(
            ({ value }: { value: string }) => useDebounce(value, 300),
            { initialProps: { value: 'a' } }
        );

        // Rapid updates
        rerender({ value: 'ab' });
        act(() => {
            vi.advanceTimersByTime(100);
        });

        rerender({ value: 'abc' });
        act(() => {
            vi.advanceTimersByTime(100);
        });

        rerender({ value: 'abcd' });

        // Still should be 'a' since debounce hasn't completed
        expect(result.current).toBe('a');

        // Complete the debounce
        act(() => {
            vi.advanceTimersByTime(300);
        });

        // Should be the last value
        expect(result.current).toBe('abcd');
    });

    it('should work with different types', () => {
        // Test with number
        const { result: numberResult } = renderHook(() => useDebounce(42, 100));
        expect(numberResult.current).toBe(42);

        // Test with object
        const obj = { key: 'value' };
        const { result: objectResult } = renderHook(() => useDebounce(obj, 100));
        expect(objectResult.current).toEqual({ key: 'value' });

        // Test with array
        const arr = [1, 2, 3];
        const { result: arrayResult } = renderHook(() => useDebounce(arr, 100));
        expect(arrayResult.current).toEqual([1, 2, 3]);
    });

    it('should use default delay of 300ms', () => {
        const { result, rerender } = renderHook(
            ({ value }: { value: string }) => useDebounce(value),
            { initialProps: { value: 'initial' } }
        );

        rerender({ value: 'updated' });

        // Should not be updated before 300ms
        act(() => {
            vi.advanceTimersByTime(200);
        });
        expect(result.current).toBe('initial');

        // Should be updated after 300ms
        act(() => {
            vi.advanceTimersByTime(100);
        });
        expect(result.current).toBe('updated');
    });

    it('should handle delay changes', () => {
        const { result, rerender } = renderHook(
            ({ value, delay }: TestProps) => useDebounce(value, delay),
            { initialProps: { value: 'initial', delay: 500 } }
        );

        // Change delay
        rerender({ value: 'updated', delay: 100 });

        // Should update after new delay
        act(() => {
            vi.advanceTimersByTime(100);
        });
        expect(result.current).toBe('updated');
    });

    it('should handle null and undefined values', () => {
        const { result: nullResult } = renderHook(() => useDebounce(null, 100));
        expect(nullResult.current).toBeNull();

        const { result: undefinedResult } = renderHook(() => useDebounce(undefined, 100));
        expect(undefinedResult.current).toBeUndefined();
    });

    it('should handle zero delay', () => {
        const { result, rerender } = renderHook(
            ({ value }: { value: string }) => useDebounce(value, 0),
            { initialProps: { value: 'initial' } }
        );

        rerender({ value: 'updated' });

        act(() => {
            vi.advanceTimersByTime(0);
        });

        expect(result.current).toBe('updated');
    });

    it('should cleanup timeout on unmount', () => {
        const clearTimeoutSpy = vi.spyOn(global, 'clearTimeout');

        const { unmount } = renderHook(() => useDebounce('value', 500));

        unmount();

        expect(clearTimeoutSpy).toHaveBeenCalled();
        clearTimeoutSpy.mockRestore();
    });
});
