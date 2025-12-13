import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

export function middleware(request: NextRequest) {
    const token = request.cookies.get('auth_token')?.value;
    const { pathname } = request.nextUrl;

    // Paths that are accessible without authentication
    const publicPaths = ['/login', '/signup', '/verify-email', '/forgot-password', '/reset-password', '/mfa'];

    // Check if the current path is public
    const isPublicPath = publicPaths.some(path => pathname.startsWith(path));

    // If user is authenticated
    if (token) {
        // Redirect to dashboard if trying to access public auth pages (login/signup) or root
        if (isPublicPath || pathname === '/') {
            return NextResponse.redirect(new URL('/dashboard', request.url));
        }
    } else {
        // If user is NOT authenticated

        // Redirect to login if trying to access protected pages
        // We treat everything as protected except public paths
        if (!isPublicPath && pathname !== '/') {
            // Allow access to static resources and api
            if (!pathname.startsWith('/_next') && !pathname.startsWith('/static') && !pathname.startsWith('/api') && !pathname.includes('.')) {
                const loginUrl = new URL('/login', request.url);
                // Add return URL only for UX (optional)
                loginUrl.searchParams.set('from', pathname);
                return NextResponse.redirect(loginUrl);
            }
        }

        // If accessing root without token, redirect to login
        if (pathname === '/') {
            return NextResponse.redirect(new URL('/login', request.url));
        }
    }

    return NextResponse.next();
}

export const config = {
    matcher: [
        /*
         * Match all request paths except for the ones starting with:
         * - api (API routes)
         * - _next/static (static files)
         * - _next/image (image optimization files)
         * - favicon.ico (favicon file)
         * - public files (images, etc)
         */
        '/((?!api|_next/static|_next/image|favicon.ico|.*\\..*).*)',
    ],
};
