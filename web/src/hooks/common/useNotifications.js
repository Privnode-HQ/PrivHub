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

import { useEffect, useState } from 'react';
import { API } from '../../helpers';

export const MESSAGE_UNREAD_REFRESH_EVENT = 'message-unread-refresh';

export const refreshUnreadMessages = () => {
  window.dispatchEvent(new Event(MESSAGE_UNREAD_REFRESH_EVENT));
};

export const useNotifications = (enabled) => {
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    const loadUnreadCount = async () => {
      if (!enabled) {
        setUnreadCount(0);
        return;
      }

      try {
        const res = await API.get('/api/message/self/unread', {
          disableDuplicate: true,
          skipErrorHandler: true,
        });
        if (res.data.success) {
          setUnreadCount(res.data.data || 0);
        }
      } catch (error) {
        setUnreadCount(0);
      }
    };

    loadUnreadCount();
    const intervalId = window.setInterval(loadUnreadCount, 60000);

    const handleRefresh = () => {
      loadUnreadCount();
    };

    window.addEventListener(MESSAGE_UNREAD_REFRESH_EVENT, handleRefresh);
    return () => {
      window.clearInterval(intervalId);
      window.removeEventListener(MESSAGE_UNREAD_REFRESH_EVENT, handleRefresh);
    };
  }, [enabled]);

  return {
    unreadCount,
  };
};
