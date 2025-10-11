import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Next.js middleware for route protection and authentication
 * Runs before page rendering on the server
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Get authentication token from cookies or session storage
  // Note: sessionStorage is not accessible in middleware, so we rely on cookies
  const accessToken = request.cookies.get('access_token')?.value;

  // Public routes that don't require authentication
  const publicRoutes = [
    '/login',
    '/signup',
    '/forgot-password',
    '/reset-password',
    '/verify-email',
    '/mfa',
    '/',
  ];

  // Check if current path is a public route
  const isPublicRoute = publicRoutes.some(route => 
    pathname === route || pathname.startsWith(`${route}/`)
  );

  // Protected routes (dashboard and its sub-routes)
  const isProtectedRoute = pathname.startsWith('/dashboard');

  // Redirect unauthenticated users from protected routes to login
  if (isProtectedRoute && !accessToken) {
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // Redirect authenticated users from auth pages to dashboard
  if (accessToken && (pathname === '/login' || pathname === '/signup')) {
    return NextResponse.redirect(new URL('/dashboard', request.url));
  }

  // Add security headers
  const response = NextResponse.next();

  // Content Security Policy
  response.headers.set(
    'Content-Security-Policy',
    "default-src 'self'; " +
    "script-src 'self' 'unsafe-eval' 'unsafe-inline'; " +
    "style-src 'self' 'unsafe-inline'; " +
    "img-src 'self' data: https:; " +
    "font-src 'self' data:; " +
    "connect-src 'self' http://localhost:8081 https:; " +
    "frame-ancestors 'none';"
  );

  // X-Frame-Options: Prevent clickjacking
  response.headers.set('X-Frame-Options', 'DENY');

  // X-Content-Type-Options: Prevent MIME sniffing
  response.headers.set('X-Content-Type-Options', 'nosniff');

  // Referrer-Policy: Control referrer information
  response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');

  // Permissions-Policy: Control browser features
  response.headers.set(
    'Permissions-Policy',
    'camera=(), microphone=(), geolocation=()'
  );

  // X-XSS-Protection: Enable XSS filter (legacy, but doesn't hurt)
  response.headers.set('X-XSS-Protection', '1; mode=block');

  return response;
}

/**
 * Configure which routes the middleware runs on
 * Excludes static files, images, and API routes
 */
export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public folder files
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
