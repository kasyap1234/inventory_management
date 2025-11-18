'use client';

import { Bell, ChevronDown, Settings, User, LogOut, Moon, Sun } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useState, useRef, useEffect } from 'react';
import Link from 'next/link';
import { cn } from '@/lib/utils';
import { useTheme } from 'next-themes';
import { Button } from '@/components/ui/button';

export function Header() {
  const { user, logout } = useAuth();
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { theme, setTheme } = useTheme();

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
    <header className="h-16 bg-background/80 backdrop-blur-md border-b border-border fixed top-0 right-0 left-64 z-20 transition-all duration-200">
      <div className="h-full px-6 flex items-center justify-between">
        {/* Left side - can add breadcrumbs or page title here */}
        <div className="flex items-center gap-4">
          <h2 className="text-lg font-semibold text-foreground">Dashboard</h2>
        </div>

        {/* Right side - Profile and notifications */}
        <div className="flex items-center gap-4">
          {/* Theme Toggle */}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            className="text-muted-foreground hover:text-foreground"
          >
            <Sun className="h-5 w-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
            <Moon className="absolute h-5 w-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
            <span className="sr-only">Toggle theme</span>
          </Button>

          {/* Notifications */}
          <button
            className="relative p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors"
            aria-label="Notifications"
          >
            <Bell className="w-5 h-5" />
            {/* Notification badge */}
            <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-destructive rounded-full"></span>
          </button>

          {/* Profile Dropdown */}
          <div className="relative" ref={dropdownRef}>
            <button
              onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              className="flex items-center gap-3 px-2 py-1.5 hover:bg-muted/50 rounded-full border border-transparent hover:border-border transition-all duration-200"
              aria-expanded={isDropdownOpen}
              aria-haspopup="true"
            >
              <div className="flex items-center justify-center w-8 h-8 bg-primary text-primary-foreground rounded-full font-medium text-xs shadow-sm">
                {user?.first_name?.[0]}{user?.last_name?.[0]}
              </div>
              <div className="hidden md:block text-left">
                <p className="text-sm font-semibold text-foreground">
                  {user?.first_name} {user?.last_name}
                </p>
                <p className="text-xs text-muted-foreground">{user?.email}</p>
              </div>
              <ChevronDown
                className={cn(
                  "w-4 h-4 text-muted-foreground transition-transform",
                  isDropdownOpen && "rotate-180"
                )}
              />
            </button>

            {/* Dropdown Menu */}
            {isDropdownOpen && (
              <div className="absolute right-0 mt-2 w-72 bg-popover border border-border rounded-xl shadow-lg py-2 animate-in fade-in zoom-in-95 duration-200 origin-top-right">
                {/* User Info */}
                <div className="px-4 py-3 border-b border-border">
                  <p className="text-sm font-semibold text-foreground">
                    {user?.first_name} {user?.last_name}
                  </p>
                  <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
                  {user?.role && (
                    <span className="inline-block mt-2 px-2 py-1 text-xs font-medium bg-primary/10 text-primary rounded-md border border-primary/20">
                      {user.role}
                    </span>
                  )}
                </div>

                {/* Menu Items */}
                <div className="py-1">
                  <Link
                    href="/dashboard/profile"
                    onClick={() => setIsDropdownOpen(false)}
                    className="flex items-center gap-3 px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                  >
                    <User className="w-4 h-4 text-muted-foreground" />
                    My Profile
                  </Link>
                  <Link
                    href="/dashboard/settings"
                    onClick={() => setIsDropdownOpen(false)}
                    className="flex items-center gap-3 px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                  >
                    <Settings className="w-4 h-4 text-muted-foreground" />
                    Settings
                  </Link>
                </div>

                {/* Logout */}
                <div className="border-t border-border py-1">
                  <button
                    onClick={() => {
                      setIsDropdownOpen(false);
                      logout.mutate();
                    }}
                    className="flex items-center gap-3 w-full px-4 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors"
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
