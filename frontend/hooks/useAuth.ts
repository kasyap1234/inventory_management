import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import api from '@/lib/api';
import { AuthResponse, LoginCredentials, SignupData, User } from '@/types';

export function useAuth() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  const login = useMutation({
    mutationFn: async (credentials: LoginCredentials) => {
      const response = await api.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    },
    onSuccess: (data) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
      }
      queryClient.setQueryData(['user'], data.user);
      router.push('/dashboard');
    },
  });

  const signup = useMutation({
    mutationFn: async (data: SignupData) => {
      const response = await api.post<AuthResponse>('/auth/signup', data);
      return response.data;
    },
    onSuccess: (data) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
      }
      queryClient.setQueryData(['user'], data.user);
      router.push('/dashboard');
    },
  });

  const logout = useMutation({
    mutationFn: async () => {
      await api.post('/auth/logout');
    },
    onSuccess: () => {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
      }
      queryClient.clear();
      router.push('/login');
    },
  });

  const hasToken = isClient && typeof window !== 'undefined' && !!localStorage.getItem('access_token');

  const { data: user, isLoading } = useQuery<User>({
    queryKey: ['user'],
    queryFn: async () => {
      const response = await api.get<User>('/me');
      return response.data;
    },
    enabled: hasToken,
    retry: false,
  });

  return {
    user,
    isLoading: !isClient || isLoading,
    isAuthenticated: !!user,
    login,
    signup,
    logout,
  };
}
