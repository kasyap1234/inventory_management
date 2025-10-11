'use client';

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { User, Lock, Building } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { useAuth } from '@/hooks/useAuth';
import { useToast } from '@/components/ui/toast';
import api from '@/lib/api';

export default function SettingsPage() {
  const { user } = useAuth();
  const { addToast } = useToast();
  const [activeTab, setActiveTab] = useState<'profile' | 'password' | 'tenant'>('profile');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="text-gray-500 mt-1">Manage your account and preferences</p>
      </div>

      <div className="flex gap-4 border-b border-gray-200">
        <button
          onClick={() => setActiveTab('profile')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'profile'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          <User className="inline h-4 w-4 mr-2" />
          Profile
        </button>
        <button
          onClick={() => setActiveTab('password')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'password'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          <Lock className="inline h-4 w-4 mr-2" />
          Password
        </button>
        <button
          onClick={() => setActiveTab('tenant')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'tenant'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-gray-900'
          }`}
        >
          <Building className="inline h-4 w-4 mr-2" />
          Organization
        </button>
      </div>

      {activeTab === 'profile' && user && <ProfileSettings user={user} />}
      {activeTab === 'password' && <PasswordSettings />}
      {activeTab === 'tenant' && <TenantSettings />}
    </div>
  );
}

interface UserProfile {
  id?: string;
  first_name?: string;
  last_name?: string;
  email?: string;
}

function ProfileSettings({ user }: { user: UserProfile }) {
  const { addToast } = useToast();
  const [formData, setFormData] = useState({
    first_name: user?.first_name || '',
    last_name: user?.last_name || '',
    email: user?.email || '',
  });

  const updateProfile = useMutation({
    mutationFn: async (data: typeof formData) => {
      const response = await api.put(`/users/${user?.id}`, data);
      return response.data;
    },
    onSuccess: () => {
      addToast('Profile updated successfully', 'success');
    },
    onError: () => {
      addToast('Failed to update profile', 'error');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateProfile.mutate(formData);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Profile Information</CardTitle>
        <CardDescription>Update your personal information</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">First Name</label>
              <Input
                value={formData.first_name}
                onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
                placeholder="John"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Last Name</label>
              <Input
                value={formData.last_name}
                onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
                placeholder="Doe"
              />
            </div>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Email</label>
            <Input
              type="email"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              placeholder="you@example.com"
            />
          </div>
          <div className="flex justify-end">
            <Button type="submit" disabled={updateProfile.isPending}>
              {updateProfile.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function PasswordSettings() {
  const { addToast } = useToast();
  const [formData, setFormData] = useState({
    current_password: '',
    new_password: '',
    confirm_password: '',
  });

  const changePassword = useMutation({
    mutationFn: async (data: { current_password: string; new_password: string }) => {
      const response = await api.post('/auth/change-password', data);
      return response.data;
    },
    onSuccess: () => {
      addToast('Password changed successfully', 'success');
      setFormData({ current_password: '', new_password: '', confirm_password: '' });
    },
    onError: () => {
      addToast('Failed to change password', 'error');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (formData.new_password !== formData.confirm_password) {
      addToast('Passwords do not match', 'error');
      return;
    }
    changePassword.mutate({
      current_password: formData.current_password,
      new_password: formData.new_password,
    });
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Change Password</CardTitle>
        <CardDescription>Update your password to keep your account secure</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Current Password</label>
            <Input
              type="password"
              value={formData.current_password}
              onChange={(e) => setFormData({ ...formData, current_password: e.target.value })}
              placeholder="••••••••"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">New Password</label>
            <Input
              type="password"
              value={formData.new_password}
              onChange={(e) => setFormData({ ...formData, new_password: e.target.value })}
              placeholder="••••••••"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Confirm New Password</label>
            <Input
              type="password"
              value={formData.confirm_password}
              onChange={(e) => setFormData({ ...formData, confirm_password: e.target.value })}
              placeholder="••••••••"
              required
            />
          </div>
          <div className="flex justify-end">
            <Button type="submit" disabled={changePassword.isPending}>
              {changePassword.isPending ? 'Changing...' : 'Change Password'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

function TenantSettings() {
  const { addToast } = useToast();

  return (
    <Card>
      <CardHeader>
        <CardTitle>Organization Settings</CardTitle>
        <CardDescription>Manage your organization details</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Organization Name</label>
            <Input value="Agromart India Pvt Ltd" disabled />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Subdomain</label>
            <Input value="agromart.inventory.com" disabled />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">License Number</label>
            <Input placeholder="Enter license number" />
          </div>
          <p className="text-sm text-gray-500">
            Contact support to update organization details
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
