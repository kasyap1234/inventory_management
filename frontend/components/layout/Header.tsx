'use client';

import { Bell, ChevronDown, Settings, User, LogOut } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { cn } from '@/lib/utils';

export function Header() {
  const { user, logout } = useAuth();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return (
    <header className="h-16 bg-white border-b border-gray-200 fixed top-0 right-0 left-64 z-10">
      <div className="h-full px-6 flex items-center justify-between">
        {/* Left side - can add breadcrumbs or page title here */}
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold text-gray-900">Dashboard</h2>
        </div>

        {/* Right side - Profile and notifications */}
        <div className="flex items-center gap-4">
          {/* Notifications */}
          <button
            className="relative p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
            aria-label="Notifications"
          >
            <Bell className="w-5 h-5" />
            {/* Notification badge */}
            <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full"></span>
          </button>

          {/* Profile Dropdown */}
          <div className="relative" ref={dropdownRef}>
            <button
              onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              className="flex items-center gap-3 px-3 py-2 hover:bg-gray-50 rounded-lg transition-colors"
              aria-expanded={isDropdownOpen}
              aria-haspopup="true"
            >
              <div className="flex items-center justify-center w-9 h-9 gradient-primary rounded-full text-white font-semibold text-sm shadow-sm">
                {user?.first_name?.[0]}{user?.last_name?.[0]}
              </div>
              <div className="hidden md:block text-left">
                <p className="text-sm font-semibold text-gray-900">
                  {user?.first_name} {user?.last_name}
                </p>
                <p className="text-xs text-gray-500">{user?.email}</p>
              </div>
              <ChevronDown
                className={cn(
                  "w-4 h-4 text-gray-500 transition-transform",
                  isDropdownOpen && "rotate-180"
                )}
              />
            </button>

            {/* Dropdown Menu */}
            {isDropdownOpen && (
              <div className="absolute right-0 mt-2 w-64 bg-white border border-gray-200 rounded-lg shadow-lg py-1 animate-minimal-scale">
                {/* User Info */}
                <div className="px-4 py-3 border-b border-gray-100">
                  <p className="text-sm font-semibold text-gray-900">
                    {user?.first_name} {user?.last_name}
                  </p>
                  <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                  {user?.role && (
                    <span className="inline-block mt-2 px-2 py-1 text-xs font-medium bg-blue-50 text-blue-700 rounded-md border border-blue-200">
                      {user.role}
                    </span>
                  )}
                </div>

                {/* Menu Items */}
                <div className="py-1">
                  <Link
                    href="/dashboard/profile"
                    onClick={() => setIsDropdownOpen(false)}
                    className="flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    <User className="w-4 h-4 text-gray-500" />
                    My Profile
                  </Link>
                  <Link
                    href="/dashboard/settings"
                    onClick={() => setIsDropdownOpen(false)}
                    className="flex items-center gap-3 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 transition-colors"
                  >
                    <Settings className="w-4 h-4 text-gray-500" />
                    Settings
                  </Link>
                </div>

                {/* Logout */}
                <div className="border-t border-gray-100 py-1">
                  <button
                    onClick={() => {
                      setIsDropdownOpen(false);
                      logout.mutate();
                    }}
                    className="flex items-center gap-3 w-full px-4 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                  >
                    <LogOut className="w-4 h-4" />
                    Logout
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
