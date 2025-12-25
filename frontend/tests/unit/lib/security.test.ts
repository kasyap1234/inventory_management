import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { csrfTokenManager, tokenStorage, mfaChallengeStore } from '@/lib/security';

// Mock sessionStorage for MFA tests
const sessionStorageMock = (() => {
    let store: Record<string, string> = {};
    return {
        getItem: (key: string) => store[key] || null,
        setItem: (key: string, value: string) => {
            store[key] = value;
        },
        removeItem: (key: string) => {
            delete store[key];
        },
        clear: () => {
            store = {};
        },
    };
})();

Object.defineProperty(window, 'sessionStorage', {
    value: sessionStorageMock,
});

describe('tokenStorage', () => {
    // TokenStorage now uses HttpOnly cookies, so these methods return null/no-op
    it('should return null for access token (HttpOnly cookie)', () => {
        expect(tokenStorage.getAccessToken()).toBeNull();
    });

    it('should return null for refresh token (HttpOnly cookie)', () => {
        expect(tokenStorage.getRefreshToken()).toBeNull();
    });

    it('should have clear method that does not throw', () => {
        expect(() => tokenStorage.clear()).not.toThrow();
    });

    it('should have setTokens method for backward compatibility', () => {
        expect(() => tokenStorage.setTokens()).not.toThrow();
    });

    it('should return false for hasAccessToken (HttpOnly cookies not accessible)', () => {
        expect(tokenStorage.hasAccessToken()).toBe(false);
    });
});

describe('mfaChallengeStore', () => {
    beforeEach(() => {
        sessionStorageMock.clear();
    });

    it('should store and retrieve MFA token', () => {
        mfaChallengeStore.store('test-mfa-token');
        expect(mfaChallengeStore.get()).toBe('test-mfa-token');
    });

    it('should clear MFA token', () => {
        mfaChallengeStore.store('test-mfa-token');
        mfaChallengeStore.clear();
        expect(mfaChallengeStore.get()).toBeNull();
    });

    it('should handle null token', () => {
        mfaChallengeStore.store(null);
        expect(mfaChallengeStore.get()).toBeNull();
    });

    it('should handle undefined token', () => {
        mfaChallengeStore.store(undefined);
        expect(mfaChallengeStore.get()).toBeNull();
    });
});

describe('csrfTokenManager', () => {
    beforeEach(() => {
        vi.resetAllMocks();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should clear token without throwing', () => {
        expect(() => csrfTokenManager.clearToken()).not.toThrow();
    });

    it('should handle network errors gracefully', async () => {
        // Mock fetch to simulate network error
        global.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

        csrfTokenManager.clearToken();
        const token = await csrfTokenManager.getToken();

        // Should return null on error
        expect(token === null || token === '' || token === undefined).toBe(true);
    });
});
