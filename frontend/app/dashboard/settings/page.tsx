'use client';

import { useState, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { User, Lock, Building, Shield } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { useAuth } from '@/hooks/useAuth';
import { useToast } from '@/components/ui/toast';
import api from '@/lib/api';

export default function SettingsPage() {
  const { user } = useAuth();
  const { addToast } = useToast();
  const [activeTab, setActiveTab] = useState<'profile' | 'password' | 'tenant' | 'security'>('profile');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground">Settings</h1>
        <p className="text-gray-500 mt-1">Manage your account and preferences</p>
      </div>

      <div className="flex gap-4 border-b border-gray-200">
        <button
          onClick={() => setActiveTab('profile')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'profile'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-foreground'
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
              : 'text-gray-600 hover:text-foreground'
          }`}
        >
          <Lock className="inline h-4 w-4 mr-2" />
          Password
        </button>
        <button
          onClick={() => setActiveTab('security')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'security'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-foreground'
          }`}
        >
          <Shield className="inline h-4 w-4 mr-2" />
          Security
        </button>
        <button
          onClick={() => setActiveTab('tenant')}
          className={`px-4 py-2 font-medium transition-colors ${
            activeTab === 'tenant'
              ? 'border-b-2 border-blue-600 text-blue-600'
              : 'text-gray-600 hover:text-foreground'
          }`}
        >
          <Building className="inline h-4 w-4 mr-2" />
          Organization
        </button>
      </div>

      {activeTab === 'profile' && user && <ProfileSettings user={user} />}
      {activeTab === 'password' && <PasswordSettings />}
      {activeTab === 'security' && <SecuritySettings />}
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
   });

   const updateProfile = useMutation({
     mutationFn: async (data: typeof formData) => {
       const response = await api.put('/users/me', data);
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
                 required
               />
             </div>
             <div className="space-y-2">
               <label className="text-sm font-medium">Last Name</label>
               <Input
                 value={formData.last_name}
                 onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
                 placeholder="Doe"
                 required
               />
             </div>
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
 
 function SecuritySettings() {
   const { addToast } = useToast();
   const [twoFactorEnabled, setTwoFactorEnabled] = useState(false);
   const [qrCodeUrl, setQrCodeUrl] = useState('');
   const [showSetup, setShowSetup] = useState(false);
   const [code, setCode] = useState('');
 
   useEffect(() => {
     // Check if 2FA is enabled
     const check2FAStatus = async () => {
       try {
         const response = await api.get('/me');
         setTwoFactorEnabled(response.data.two_factor_enabled || false);
       } catch (error) {
         console.error('Failed to check 2FA status:', error);
       }
     };
     check2FAStatus();
   }, []);
 
   const generate2FA = useMutation({
     mutationFn: async () => {
       const response = await api.post('/auth/2fa/generate');
       return response.data;
     },
     onSuccess: (data) => {
       setQrCodeUrl(data.qr_code_url);
       setShowSetup(true);
     },
     onError: () => {
       addToast('Failed to generate 2FA secret', 'error');
     },
   });
 
   const enable2FA = useMutation({
     mutationFn: async (code: string) => {
       const response = await api.post('/auth/2fa/enable', { code });
       return response.data;
     },
     onSuccess: () => {
       setTwoFactorEnabled(true);
       setShowSetup(false);
       setQrCodeUrl('');
       setCode('');
       addToast('2FA enabled successfully', 'success');
     },
     onError: () => {
       addToast('Invalid 2FA code', 'error');
     },
   });
 
   const disable2FA = useMutation({
     mutationFn: async () => {
       const response = await api.post('/auth/2fa/disable');
       return response.data;
     },
     onSuccess: () => {
       setTwoFactorEnabled(false);
       addToast('2FA disabled successfully', 'success');
     },
     onError: () => {
       addToast('Failed to disable 2FA', 'error');
     },
   });
 
   const handleEnable2FA = () => {
     generate2FA.mutate();
   };
 
   const handleConfirmEnable = (e: React.FormEvent) => {
     e.preventDefault();
     if (code.length === 6) {
       enable2FA.mutate(code);
     } else {
       addToast('Please enter a valid 6-digit code', 'error');
     }
   };
 
   const handleDisable2FA = () => {
     if (confirm('Are you sure you want to disable 2FA? This will make your account less secure.')) {
       disable2FA.mutate();
     }
   };
 
   return (
     <Card>
       <CardHeader>
         <CardTitle>Two-Factor Authentication</CardTitle>
         <CardDescription>Add an extra layer of security to your account</CardDescription>
       </CardHeader>
       <CardContent className="space-y-6">
         {twoFactorEnabled ? (
           <div className="space-y-4">
             <div className="flex items-center space-x-2">
               <Shield className="h-5 w-5 text-green-600" />
               <span className="text-sm font-medium text-green-600">2FA is enabled</span>
             </div>
             <p className="text-sm text-gray-600">
               Your account is protected with two-factor authentication. You'll need your authenticator app to sign in.
             </p>
             <Button
               variant="outline"
               onClick={handleDisable2FA}
               disabled={disable2FA.isPending}
               className="text-red-600 border-red-300 hover:bg-red-50"
             >
               {disable2FA.isPending ? 'Disabling...' : 'Disable 2FA'}
             </Button>
           </div>
         ) : (
           <div className="space-y-4">
             <div className="flex items-center space-x-2">
               <Shield className="h-5 w-5 text-gray-400" />
               <span className="text-sm font-medium text-gray-600">2FA is not enabled</span>
             </div>
             <p className="text-sm text-gray-600">
               Add an extra layer of security to your account by enabling two-factor authentication.
             </p>
             {!showSetup ? (
               <Button onClick={handleEnable2FA} disabled={generate2FA.isPending}>
                 {generate2FA.isPending ? 'Setting up...' : 'Enable 2FA'}
               </Button>
             ) : (
               <div className="space-y-4">
                 <div className="text-center">
                   <p className="text-sm text-gray-600 mb-4">
                     Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
                   </p>
                   {qrCodeUrl && (
                     <div className="inline-block p-4 border border-gray-200 rounded">
                       <img src={`https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(qrCodeUrl)}`} alt="QR Code" className="w-48 h-48" />
                     </div>
                   )}
                 </div>
                 <form onSubmit={handleConfirmEnable} className="space-y-4">
                   <div className="space-y-2">
                     <label className="text-sm font-medium">Enter the 6-digit code from your app</label>
                     <Input
                       type="text"
                       value={code}
                       onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                       placeholder="000000"
                       maxLength={6}
                       required
                       className="text-center text-lg tracking-widest"
                     />
                   </div>
                   <div className="flex gap-2">
                     <Button type="submit" disabled={enable2FA.isPending || code.length !== 6}>
                       {enable2FA.isPending ? 'Enabling...' : 'Enable 2FA'}
                     </Button>
                     <Button
                       type="button"
                       variant="outline"
                       onClick={() => {
                         setShowSetup(false);
                         setQrCodeUrl('');
                         setCode('');
                       }}
                     >
                       Cancel
                     </Button>
                   </div>
                 </form>
               </div>
             )}
           </div>
         )}
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

interface TenantSettingsData {
  name: string;
  subdomain: string;
  license: string;
}

function TenantSettings() {
  const { addToast } = useToast();
  const [formData, setFormData] = useState<TenantSettingsData>({
    name: '',
    subdomain: '',
    license: '',
  });
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchTenantSettings = async () => {
      try {
        const response = await api.get('/tenant/settings');
        const data = response.data;
        setFormData({
          name: data.name || '',
          subdomain: data.subdomain || '',
          license: data.license || '',
        });
      } catch (error) {
        addToast('Failed to load tenant settings', 'error');
      } finally {
        setIsLoading(false);
      }
    };

    fetchTenantSettings();
  }, [addToast]);

  const updateSettings = useMutation({
    mutationFn: async (data: TenantSettingsData) => {
      const response = await api.put('/tenant/settings', data);
      return response.data;
    },
    onSuccess: () => {
      addToast('Settings updated successfully', 'success');
    },
    onError: () => {
      addToast('Failed to update settings', 'error');
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    updateSettings.mutate(formData);
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Organization Settings</CardTitle>
          <CardDescription>Manage your organization details</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-gray-500">Loading...</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Organization Settings</CardTitle>
        <CardDescription>Manage your organization details</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Organization Name</label>
            <Input
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="Enter organization name"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Subdomain</label>
            <Input
              value={formData.subdomain}
              onChange={(e) => setFormData({ ...formData, subdomain: e.target.value })}
              placeholder="Enter subdomain"
              required
            />
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">License Number</label>
            <Input
              value={formData.license}
              onChange={(e) => setFormData({ ...formData, license: e.target.value })}
              placeholder="Enter license number"
            />
          </div>
          <div className="flex justify-end">
            <Button type="submit" disabled={updateSettings.isPending}>
              {updateSettings.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
