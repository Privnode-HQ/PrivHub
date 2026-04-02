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
import { Button, Space, Tag } from '@douyinfe/semi-ui';
import { formatCurrencyAmountByCode, timestamp2string } from '../../../helpers';
import { TOPUP_COUPON_STATUS_MAP } from '../../../constants/topup-coupon.constants';

const getDisplayStatus = (record) => {
  return record?.effective_status || record?.status || '';
};

const renderStatus = (status, t) => {
  const statusConfig = TOPUP_COUPON_STATUS_MAP[status];
  if (!statusConfig) {
    return (
      <Tag color='black' shape='circle'>
        {t('未知状态')}
      </Tag>
    );
  }

  return (
    <Tag color={statusConfig.color} shape='circle'>
      {t(statusConfig.text)}
    </Tag>
  );
};

const renderTime = (timestamp, emptyText, t) => {
  if (!timestamp) {
    return t(emptyText);
  }
  return timestamp2string(timestamp);
};

export const getTopupCouponsColumns = ({
  t,
  openEdit,
  openRevoke,
  readOnlyAdmin,
}) => {
  const columns = [
    {
      title: t('ID'),
      dataIndex: 'id',
    },
    {
      title: t('名称'),
      dataIndex: 'name',
    },
    {
      title: t('用户'),
      dataIndex: 'bound_username',
      render: (text, record) => {
        return (
          <div>
            {text || t('用户')} #{record.bound_user_id}
          </div>
        );
      },
    },
    {
      title: t('优惠金额'),
      dataIndex: 'deduction_amount',
      render: (text, record) => {
        return (
          <Tag color='green' shape='circle'>
            - {formatCurrencyAmountByCode(text, record.currency_code)}
          </Tag>
        );
      },
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (text, record) => renderStatus(getDisplayStatus(record), t),
    },
    {
      title: t('生效时间'),
      dataIndex: 'valid_from',
      render: (text) => renderTime(text, '立即生效', t),
    },
    {
      title: t('过期时间'),
      dataIndex: 'expires_at',
      render: (text) => renderTime(text, '永不过期', t),
    },
    {
      title: t('发放时间'),
      dataIndex: 'issued_at',
      render: (text) => renderTime(text, '无', t),
    },
  ];

  if (!readOnlyAdmin) {
    columns.push({
      title: '',
      dataIndex: 'operate',
      fixed: 'right',
      width: 180,
      render: (text, record) => {
        const displayStatus = getDisplayStatus(record);
        const canEdit =
          displayStatus !== 'used' && displayStatus !== 'reserved';
        const canRevoke =
          displayStatus === 'available' || displayStatus === 'expired';

        return (
          <Space>
            <Button
              type='tertiary'
              size='small'
              onClick={() => openEdit(record)}
              disabled={!canEdit}
            >
              {t('编辑')}
            </Button>
            <Button
              type='danger'
              size='small'
              onClick={() => openRevoke(record)}
              disabled={!canRevoke}
            >
              {t('撤销')}
            </Button>
          </Space>
        );
      },
    });
  }

  return columns;
};
