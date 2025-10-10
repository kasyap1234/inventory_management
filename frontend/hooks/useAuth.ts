import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { csrfTokenManager, mfaChallengeStore, tokenStorage } from '@/lib/security';
import {
  AuthResponse,
  ForgotPasswordPayload,
  LoginCredentials,
  ResetPasswordPayload,
  SignupData,
  SignupResponse,
  User,
} from '@/types';

export function useAuth() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const login = useMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      const payload: LoginCredentials = {
        email: credentials.email.trim().toLowerCase(),
        password: credentials.password.trim(),
      };
      const response = await api.post<AuthResponse>('/auth/login', payload);
      return response.data;
    },
    onSuccess: (data) => {
      if (data.mfa_required) {
        mfaChallengeStore.store(data.mfa_token ?? null);
        tokenStorage.clear();
        queryClient.removeQueries({ queryKey: ['user'] });
        router.push('/mfa');
        return;
      }

      tokenStorage.setTokens(data.access_token, data.refresh_token);
      csrfTokenManager.clearToken();
      queryClient.setQueryData(['user'], data.user);
      router.push('/dashboard');
    },
  });

  const signup = useMutation({
    mutationFn: async (data: SignupData) => {
      const payload: SignupData = {
        email: data.email.trim().toLowerCase(),
        password: data.password.trim(),
        first_name: data.first_name.trim(),
        last_name: data.last_name.trim(),
        tenant_name: data.tenant_name.trim(),
        subdomain: data.subdomain.trim().toLowerCase(),
      };
      const response = await api.post<SignupResponse>('/auth/signup', payload);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.removeQueries({ queryKey: ['user'] });
      const email = data.user?.email ?? '';
      const target = email ? `/verify-email?email=${encodeURIComponent(email)}` : '/verify-email';
      router.push(target);
    },
  });

  const logout = useMutation({
    mutationFn: async () => {
      await api.post('/auth/logout');
    },
    onSuccess: () => {
      tokenStorage.clear();
      csrfTokenManager.clearToken();
      mfaChallengeStore.clear();
      queryClient.clear();
      router.push('/login');
    },
  });

  const requestPasswordReset = useMutation({
    mutationFn: async (payload: ForgotPasswordPayload) => {
      const sanitized: ForgotPasswordPayload = {
        email: payload.email.trim().toLowerCase(),
      };
      await api.post('/auth/password/forgot', sanitized);
    },
  });

  const resetPassword = useMutation({
    mutationFn: async (payload: ResetPasswordPayload) => {
      const sanitized: ResetPasswordPayload = {
        token: payload.token.trim(),
        password: payload.password.trim(),
        confirm_password: payload.confirm_password.trim(),
      };
      await api.post('/auth/password/reset', sanitized);
    },
    onSuccess: () => {
      router.push('/login?reset=success');
    },
  });

  const isAuthenticatedToken = isClient && tokenStorage.hasAccessToken();

  const { data: user, isLoading } = useQuery<User>({
    queryKey: ['user'],
    queryFn: async () => {
      const response = await api.get<User>('/me');
      return response.data;
    },
    enabled: isAuthenticatedToken,
    retry: false,
  });

  return {
    user,
    isLoading: !isClient || isLoading,
    isAuthenticated: !!user,
    login,
    signup,
    logout,
    requestPasswordReset,
    resetPassword,
  };
}
