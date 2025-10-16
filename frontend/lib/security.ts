import axios, { AxiosInstance } from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

class TokenStorage {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;

  constructor() {
    if (typeof window !== 'undefined') {
      this.accessToken = sessionStorage.getItem('access_token');
      this.refreshToken = sessionStorage.getItem('refresh_token');
    }
  }

  getAccessToken() {
    if (typeof window === 'undefined') return null;
    if (!this.accessToken) {
      this.accessToken = sessionStorage.getItem('access_token');
    }
    return this.accessToken;
  }

  getRefreshToken() {
    if (typeof window === 'undefined') return null;
    if (!this.refreshToken) {
      this.refreshToken = sessionStorage.getItem('refresh_token');
    }
    return this.refreshToken;
  }

  setTokens(accessToken?: string, refreshToken?: string) {
    if (typeof window === 'undefined') return;

    if (accessToken) {
      this.accessToken = accessToken;
      sessionStorage.setItem('access_token', accessToken);
    }

    if (refreshToken) {
      this.refreshToken = refreshToken;
      sessionStorage.setItem('refresh_token', refreshToken);
    }
  }

  clear() {
    if (typeof window === 'undefined') return;

    this.accessToken = null;
    this.refreshToken = null;
    sessionStorage.removeItem('access_token');
    sessionStorage.removeItem('refresh_token');
  }

  hasAccessToken() {
    return !!this.getAccessToken();
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
