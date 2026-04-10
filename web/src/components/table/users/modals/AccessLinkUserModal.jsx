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
import {
  Banner,
  Button,
  Input,
  Modal,
  Space,
  Typography,
} from '@douyinfe/semi-ui';

const { Text } = Typography;

const formatDateTime = (value) => {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};

const AccessLinkUserModal = ({
  visible,
  user,
  loading,
  generatedLink,
  expiresAt,
  onCancel,
  onGenerate,
  onCopy,
  t,
}) => {
  return (
    <Modal
      title={t('一次性访问链接')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={620}
    >
      <div className='flex flex-col gap-4'>
        <div>
          <Text>
            {t('目标用户')}: <Text strong>{user?.username || '-'}</Text>
          </Text>
        </div>

        <Banner
          type='warning'
          title={t('完整权限访问')}
          description={t(
            '生成后会创建一个只能使用一次的账户访问链接。链接需在 24 小时内访问，访问后会直接以该用户身份登录并创建一个 24 小时有效的完整权限会话。系统会立即通知该用户，并在链接被实际使用时再次通知。',
          )}
        />

        {generatedLink ? (
          <div className='rounded-2xl border border-semi-color-border p-4'>
            <div className='mb-3 flex flex-col gap-1'>
              <Text strong>{t('访问链接')}</Text>
              <Text type='secondary'>
                {t('有效期至')}: {formatDateTime(expiresAt)}
              </Text>
            </div>
            <Input value={generatedLink} readOnly />
            <Space
              style={{
                justifyContent: 'flex-end',
                width: '100%',
                marginTop: 12,
              }}
            >
              <Button onClick={onGenerate} loading={loading}>
                {t('重新生成')}
              </Button>
              <Button type='primary' onClick={onCopy}>
                {t('复制链接')}
              </Button>
            </Space>
          </div>
        ) : null}

        <Space style={{ justifyContent: 'flex-end', width: '100%' }}>
          <Button onClick={onCancel}>{t('关闭')}</Button>
          {!generatedLink ? (
            <Button type='primary' loading={loading} onClick={onGenerate}>
              {t('生成访问链接')}
            </Button>
          ) : null}
        </Space>
      </div>
    </Modal>
  );
};

export default AccessLinkUserModal;
