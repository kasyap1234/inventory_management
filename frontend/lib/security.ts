import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api/v1';

// TokenStorage is no longer needed for HttpOnly cookie authentication
// Tokens are stored in HttpOnly cookies and sent automatically with credentials: 'include'
class TokenStorage {
  // Keep these methods for backward compatibility but they no longer interact with localStorage
  // The backend sets HttpOnly cookies that are automatically sent with requests

  getAccessToken() {
    // Tokens are in HttpOnly cookies - not accessible from JavaScript
    return null;
  }

  getRefreshToken() {
    // Tokens are in HttpOnly cookies - not accessible from JavaScript
    return null;
  }

  setTokens() {
    // No-op: Tokens are set by backend as HttpOnly cookies
    // This method is kept for backward compatibility
  }

  clear() {
    // Tokens are cleared by calling the logout endpoint
    // The backend will clear the HttpOnly cookies
  }

  hasAccessToken() {
    // Cannot check token presence from JavaScript with HttpOnly cookies
    // Auth state is determined by making an authenticated API call
    return false;
  }
}

class CSRFTokenManager {
  private token: string | null = null;
  private expiresAt = 0;
  private inFlight: Promise<string | null> | null = null;
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  async getToken(): Promise<string | null> {
    if (typeof window === 'undefined') {
      return null;
    }

    const now = Date.now();
    if (this.token && this.expiresAt - now > 30_000) {
      return this.token;
    }

    if (this.inFlight) {
      return this.inFlight;
    }

    this.inFlight = this.client
      .get<{ token: string; expires_at?: string }>('/security/csrf')
      .then((response) => {
        const token = response.data.token;
        const expiresAt = response.data.expires_at ? new Date(response.data.expires_at).getTime() : now + 60 * 60 * 1000;
        this.token = token;
        this.expiresAt = expiresAt;
        return token;
      })
      .catch(() => {
        this.token = null;
        this.expiresAt = 0;
        return null;
      })
      .finally(() => {
        this.inFlight = null;
      });

    return this.inFlight;
  }

  clearToken() {
    this.token = null;
    this.expiresAt = 0;
  }
}

class MFAChallengeStore {
  private readonly storageKey = 'mfa_challenge_token';

  store(challenge: string | undefined | null) {
    if (typeof window === 'undefined') return;
    if (challenge) {
      sessionStorage.setItem(this.storageKey, challenge);
    }
  }

  get(): string | null {
    if (typeof window === 'undefined') return null;
    return sessionStorage.getItem(this.storageKey);
  }

  clear() {
    if (typeof window === 'undefined') return;
    sessionStorage.removeItem(this.storageKey);
  }
}

export const tokenStorage = new TokenStorage();
export const csrfTokenManager = new CSRFTokenManager();
export const mfaChallengeStore = new MFAChallengeStore();
