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
  const user = localStorage.getItem('user');

  if (user) {
    return <Navigate to='/console' replace />;
  }

  return children;
};

export function getRequiredUserActions(user) {
  if (!user || !Array.isArray(user.required_actions)) {
    return [];
  }
  return user.required_actions.filter(Boolean);
}

function shouldRedirectToPersonal(user, pathname) {
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
