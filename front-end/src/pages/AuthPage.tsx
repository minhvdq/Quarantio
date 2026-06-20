import React from 'react';
import { TENANT_URL } from '../config';

interface AuthPageProps {
  inviteToken?: string;
}

export function AuthPage({ inviteToken }: AuthPageProps) {
  const inviteSuffix = inviteToken ? `?invite=${encodeURIComponent(inviteToken)}` : '';
  const googleHref = `${TENANT_URL}/auth/google/login${inviteSuffix}`;
  // const microsoftHref = `${TENANT_URL}/auth/microsoft/login${inviteSuffix}`;

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 p-6">
      <div className="w-full max-w-sm bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden">
        <div className="px-8 pt-8 pb-6 text-center border-b border-gray-100">
          <div className="flex items-center justify-center gap-2 mb-1">
            <img src="/logo.png" className="w-9 h-9 object-contain" alt="Quarantio" />
            <span className="text-xl font-bold text-gray-900 tracking-tight">Quarantio</span>
          </div>
          <p className="text-xs text-gray-400 mt-1">AI-powered email compliance</p>
        </div>
        <div className="px-8 py-6 flex flex-col gap-3">
          <p className="text-sm text-gray-500 text-center m-0">
            Sign in to continue. Your account will be
            <br />
            created automatically on first sign-in.
          </p>
          {inviteToken && (
            <div className="text-sm bg-blue-50 text-blue-700 border border-blue-200 rounded-lg px-3 py-2">
              You've been invited. Sign in to accept.
            </div>
          )}
          <a
            href={googleHref}
            className="flex items-center justify-center gap-3 border border-gray-200 rounded-xl px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            style={{ textDecoration: 'none' }}
          >
            <GoogleIcon />
            Continue with Google
          </a>
          {/* Microsoft SSO — coming soon
          <a
            href={microsoftHref}
            className="flex items-center justify-center gap-3 border border-gray-200 rounded-xl px-4 py-3 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            style={{ textDecoration: 'none' }}
          >
            <MicrosoftIcon />
            Continue with Microsoft
          </a>
          */}
          <p className="text-[11px] text-gray-400 text-center m-0">
            By signing in you agree to our{' '}
            <a href="/privacy" className="text-gray-500 hover:text-gray-700" style={{ textDecoration: 'underline' }}>
              Privacy Policy
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 48 48">
      <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/>
      <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/>
      <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/>
      <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.18 1.48-4.97 2.31-8.16 2.31-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/>
    </svg>
  );
}

function MicrosoftIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 21 21">
      <rect x="1" y="1" width="9" height="9" fill="#f25022"/>
      <rect x="11" y="1" width="9" height="9" fill="#7fba00"/>
      <rect x="1" y="11" width="9" height="9" fill="#00a4ef"/>
      <rect x="11" y="11" width="9" height="9" fill="#ffb900"/>
    </svg>
  );
}
