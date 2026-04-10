/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { history } from './history';

const AUTH_REDIRECT_STORAGE_KEY = 'auth_redirect_target';

export function authHeader() {
  // return authorization header with jwt token
  let user = JSON.parse(localStorage.getItem('user'));

  if (user && user.token) {
    return { Authorization: 'Bearer ' + user.token };
  } else {
    return {};
  }
}

export const AuthRedirect = ({ children }) => {
  const location = useLocation();
  const user = localStorage.getItem('user');

  if (user) {
    const redirectTarget = getAuthRedirectTargetFromLocation(location);
    return <Navigate to={redirectTarget || '/console'} replace />;
  }

  return children;
};

export function sanitizeAuthRedirectTarget(target) {
  if (!target || typeof target !== 'string') {
    return null;
  }

  const trimmed = target.trim();

  if (!trimmed.startsWith('/')) {
    return null;
  }
  if (trimmed.startsWith('//')) {
    return null;
  }
  if (trimmed.includes('://')) {
    return null;
  }
  if (trimmed.includes('\n') || trimmed.includes('\r')) {
    return null;
  }
  if (trimmed.startsWith('/login')) {
    return null;
  }
  if (trimmed.startsWith('/api')) {
    return null;
  }

  return trimmed;
}

export function getAuthRedirectTargetFromLocation(location) {
  if (!location) {
    return null;
  }

  const searchParams = new URLSearchParams(location.search || '');
  const redirectParam = sanitizeAuthRedirectTarget(
    searchParams.get('redirect'),
  );
  if (redirectParam) {
    return redirectParam;
  }

  const from = location?.state?.from;
  const fromPath =
    from && typeof from === 'object'
      ? `${from.pathname || ''}${from.search || ''}${from.hash || ''}`
      : null;
  return sanitizeAuthRedirectTarget(fromPath);
}

export function persistPendingAuthRedirectTarget(target) {
  const sanitized = sanitizeAuthRedirectTarget(target);
  if (!sanitized) {
    sessionStorage.removeItem(AUTH_REDIRECT_STORAGE_KEY);
    return null;
  }
  sessionStorage.setItem(AUTH_REDIRECT_STORAGE_KEY, sanitized);
  return sanitized;
}

export function getPendingAuthRedirectTarget() {
  return sanitizeAuthRedirectTarget(
    sessionStorage.getItem(AUTH_REDIRECT_STORAGE_KEY),
  );
}

export function clearPendingAuthRedirectTarget() {
  sessionStorage.removeItem(AUTH_REDIRECT_STORAGE_KEY);
}

export function consumePendingAuthRedirectTarget(...candidates) {
  for (const candidate of candidates) {
    const sanitized = sanitizeAuthRedirectTarget(candidate);
    if (sanitized) {
      clearPendingAuthRedirectTarget();
      return sanitized;
    }
  }

  const storedTarget = getPendingAuthRedirectTarget();
  clearPendingAuthRedirectTarget();
  return storedTarget;
}

export function getRequiredUserActions(user) {
  if (!user || !Array.isArray(user.required_actions)) {
    return [];
  }
  return user.required_actions.filter(Boolean);
}

function shouldRedirectToPersonal(user, pathname) {
  if (user?.impersonation?.active || user?.access_link_session?.active) {
    return false;
  }
  const actions = getRequiredUserActions(user);
  if (actions.length === 0) {
    return false;
  }
  return pathname !== '/console/personal';
}

function PrivateRoute({ children }) {
  const location = useLocation();
  const raw = localStorage.getItem('user');
  if (!raw) {
    return <Navigate to='/login' state={{ from: history.location }} />;
  }
  try {
    const user = JSON.parse(raw);
    if (shouldRedirectToPersonal(user, location.pathname)) {
      return (
        <Navigate to='/console/personal' replace state={{ from: location }} />
      );
    }
  } catch (e) {
    return <Navigate to='/login' state={{ from: history.location }} />;
  }
  return children;
}

export function SupportRoute({ children }) {
  const location = useLocation();
  const raw = localStorage.getItem('user');
  if (!raw) {
    return <Navigate to='/login' state={{ from: history.location }} />;
  }
  try {
    const user = JSON.parse(raw);
    if (user && typeof user.role === 'number' && user.role >= 5) {
      if (shouldRedirectToPersonal(user, location.pathname)) {
        return (
          <Navigate to='/console/personal' replace state={{ from: location }} />
        );
      }
      return children;
    }
  } catch (e) {
    // ignore
  }
  return <Navigate to='/forbidden' replace />;
}

export function AdminRoute({ children }) {
  const location = useLocation();
  const raw = localStorage.getItem('user');
  if (!raw) {
    return <Navigate to='/login' state={{ from: history.location }} />;
  }
  try {
    const user = JSON.parse(raw);
    if (user && typeof user.role === 'number' && user.role >= 10) {
      if (shouldRedirectToPersonal(user, location.pathname)) {
        return (
          <Navigate to='/console/personal' replace state={{ from: location }} />
        );
      }
      return children;
    }
  } catch (e) {
    // ignore
  }
  return <Navigate to='/forbidden' replace />;
}

export { PrivateRoute };
