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

export const TOPUP_COUPON_STATUS = {
  AVAILABLE: 'available',
  RESERVED: 'reserved',
  USED: 'used',
  EXPIRED: 'expired',
  REVOKED: 'revoked',
};

export const TOPUP_COUPON_STATUS_MAP = {
  [TOPUP_COUPON_STATUS.AVAILABLE]: {
    color: 'green',
    text: '可用',
  },
  [TOPUP_COUPON_STATUS.RESERVED]: {
    color: 'blue',
    text: '占用中',
  },
  [TOPUP_COUPON_STATUS.USED]: {
    color: 'grey',
    text: '已使用',
  },
  [TOPUP_COUPON_STATUS.EXPIRED]: {
    color: 'orange',
    text: '已过期',
  },
  [TOPUP_COUPON_STATUS.REVOKED]: {
    color: 'red',
    text: '已撤销',
  },
};

export const TOPUP_COUPON_ACTIONS = {
  REVOKE: 'revoke',
};
