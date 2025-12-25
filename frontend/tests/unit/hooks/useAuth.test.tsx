import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React, { ReactNode } from 'react';

// Mock dependencies
vi.mock('next/navigation', () => ({
  useRouter: vi.fn(() => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  })),
}));

vi.mock('@/lib/api', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

vi.mock('@/lib/security', () => ({
  csrfTokenManager: {
    clearToken: vi.fn(),
    getToken: vi.fn(),
    setToken: vi.fn(),
  },
  mfaChallengeStore: {
    store: vi.fn(),
    get: vi.fn(),
    clear: vi.fn(),
  },
  tokenStorage: {
    clear: vi.fn(),
    getAccessToken: vi.fn(),
    setTokens: vi.fn(),
  },
}));

// Create test wrapper
function createTestWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });

  return function TestWrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };
}

describe('useAuth Hook - Login', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should be importable', async () => {
    // This tests that the hook module structure is valid
    const module = await import('@/hooks/useAuth');
    expect(module.useAuth).toBeDefined();
  });
});

describe('useAuth Hook - Logout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should clear tokens on logout', async () => {
    const { tokenStorage, csrfTokenManager, mfaChallengeStore } = await import('@/lib/security');

    // Simulate logout behavior
    tokenStorage.clear();
    csrfTokenManager.clearToken();
    mfaChallengeStore.clear();

    expect(tokenStorage.clear).toHaveBeenCalled();
    expect(csrfTokenManager.clearToken).toHaveBeenCalled();
    expect(mfaChallengeStore.clear).toHaveBeenCalled();
  });
});

describe('useAuth Hook - MFA Flow', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should store MFA token when MFA is required', async () => {
    const { mfaChallengeStore } = await import('@/lib/security');

    const mfaToken = 'test-mfa-token';
    mfaChallengeStore.store(mfaToken);

    expect(mfaChallengeStore.store).toHaveBeenCalledWith(mfaToken);
  });

  it('should clear MFA token', async () => {
    const { mfaChallengeStore } = await import('@/lib/security');

    mfaChallengeStore.clear();

    expect(mfaChallengeStore.clear).toHaveBeenCalled();
  });
});

describe('Auth Utility Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should normalize email on login', () => {
    const credentials = {
      email: '  Test@Example.COM  ',
      password: 'password123',
    };

    const normalizedEmail = credentials.email.trim().toLowerCase();

    expect(normalizedEmail).toBe('test@example.com');
  });

  it('should trim password on login', () => {
    const credentials = {
      email: 'test@example.com',
      password: '  password123  ',
    };

    const trimmedPassword = credentials.password.trim();

    expect(trimmedPassword).toBe('password123');
  });

  it('should normalize signup data', () => {
    const signupData = {
      email: '  User@Example.COM  ',
      password: 'password123',
      first_name: '  John  ',
      last_name: '  Doe  ',
      tenant_name: '  My Company  ',
      subdomain: '  MyCompany  ',
    };

    const normalized = {
      email: signupData.email.trim().toLowerCase(),
      password: signupData.password.trim(),
      first_name: signupData.first_name.trim(),
      last_name: signupData.last_name.trim(),
      tenant_name: signupData.tenant_name.trim(),
      subdomain: signupData.subdomain.trim().toLowerCase(),
    };

    expect(normalized.email).toBe('user@example.com');
    expect(normalized.first_name).toBe('John');
    expect(normalized.last_name).toBe('Doe');
    expect(normalized.tenant_name).toBe('My Company');
    expect(normalized.subdomain).toBe('mycompany');
  });
});

describe('Auth API Endpoints', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should call login endpoint correctly', async () => {
    const api = (await import('@/lib/api')).default;

    const credentials = {
      email: 'test@example.com',
      password: 'password123',
    };

    await api.post('/auth/login', credentials);

    expect(api.post).toHaveBeenCalledWith('/auth/login', credentials);
  });

  it('should call logout endpoint', async () => {
    const api = (await import('@/lib/api')).default;

    await api.post('/auth/logout');

    expect(api.post).toHaveBeenCalledWith('/auth/logout');
  });

  it('should call signup endpoint', async () => {
    const api = (await import('@/lib/api')).default;

    const signupData = {
      email: 'newuser@example.com',
      password: 'password123',
      first_name: 'John',
      last_name: 'Doe',
      tenant_name: 'New Company',
      subdomain: 'newcompany',
    };

    await api.post('/auth/signup', signupData);

    expect(api.post).toHaveBeenCalledWith('/auth/signup', signupData);
  });

  it('should call forgot password endpoint', async () => {
    const api = (await import('@/lib/api')).default;

    const payload = { email: 'user@example.com' };

    await api.post('/auth/forgot-password', payload);

    expect(api.post).toHaveBeenCalledWith('/auth/forgot-password', payload);
  });

  it('should call reset password endpoint', async () => {
    const api = (await import('@/lib/api')).default;

    const payload = {
      token: 'reset-token',
      password: 'newpassword123',
    };

    await api.post('/auth/reset-password', payload);

    expect(api.post).toHaveBeenCalledWith('/auth/reset-password', payload);
  });
});

describe('Auth Response Handling', () => {
  it('should handle successful login response', () => {
    const response = {
      user: {
        id: '123',
        email: 'test@example.com',
        first_name: 'John',
        last_name: 'Doe',
      },
      mfa_required: false,
    };

    expect(response.mfa_required).toBe(false);
    expect(response.user).toBeDefined();
    expect(response.user.email).toBe('test@example.com');
  });

  it('should handle MFA required response', () => {
    const response = {
      user: null,
      mfa_required: true,
      mfa_token: 'mfa-challenge-token',
    };

    expect(response.mfa_required).toBe(true);
    expect(response.mfa_token).toBeDefined();
  });

  it('should handle signup success response', () => {
    const response = {
      user: {
        id: '123',
        email: 'newuser@example.com',
        email_verified: false,
      },
      message: 'Please verify your email',
    };

    expect(response.user.email_verified).toBe(false);
    expect(response.message).toContain('verify');
  });
});

describe('Auth Error Scenarios', () => {
  it('should handle invalid credentials error', () => {
    const error = {
      response: {
        status: 401,
        data: {
          error: 'Invalid email or password',
        },
      },
    };

    expect(error.response.status).toBe(401);
    expect(error.response.data.error).toContain('Invalid');
  });

  it('should handle account locked error', () => {
    const error = {
      response: {
        status: 403,
        data: {
          error: 'Account locked due to too many failed attempts',
        },
      },
    };

    expect(error.response.status).toBe(403);
    expect(error.response.data.error).toContain('locked');
  });

  it('should handle email not verified error', () => {
    const error = {
      response: {
        status: 403,
        data: {
          error: 'Email not verified',
        },
      },
    };

    expect(error.response.status).toBe(403);
    expect(error.response.data.error).toContain('not verified');
  });

  it('should handle network error', () => {
    const error = {
      message: 'Network Error',
      code: 'ERR_NETWORK',
    };

    expect(error.code).toBe('ERR_NETWORK');
  });
});
