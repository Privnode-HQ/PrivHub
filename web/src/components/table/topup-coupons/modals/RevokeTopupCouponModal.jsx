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

import React, { useEffect, useState } from 'react';
import { Button, Input, Modal, Space, Typography } from '@douyinfe/semi-ui';

const { Text } = Typography;

const RevokeTopupCouponModal = ({
  visible,
  coupon,
  onCancel,
  onConfirm,
  loading,
  t,
}) => {
  const [reason, setReason] = useState('');

  useEffect(() => {
    if (!visible) {
      setReason('');
    }
  }, [visible]);

  return (
    <Modal
      title={t('撤销优惠券')}
      visible={visible}
      onCancel={onCancel}
      footer={
        <Space>
          <Button type='tertiary' onClick={onCancel}>
            {t('取消')}
          </Button>
          <Button
            type='danger'
            loading={loading}
            onClick={() => onConfirm(reason)}
          >
            {t('确认撤销')}
          </Button>
        </Space>
      }
    >
      <div className='space-y-4'>
        <Text>
          {t('将撤销优惠券')} #{coupon?.id} {coupon?.name || ''}
        </Text>
        <Input
          value={reason}
          onChange={setReason}
          placeholder={t('请输入撤销原因（可选）')}
          maxLength={255}
        />
      </div>
    </Modal>
  );
};

export default RevokeTopupCouponModal;
