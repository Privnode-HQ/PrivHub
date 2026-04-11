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

import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Descriptions,
  Empty,
  Input,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { KeyRound, UserSearch } from 'lucide-react';
import CardPro from '../../components/common/ui/CardPro';
import { API, renderQuota, showError, timestamp2string } from '../../helpers';

const { Text } = Typography;

const renderUserRole = (role, t) => {
  switch (role) {
    case 1:
      return <Tag color='blue'>{t('普通用户')}</Tag>;
    case 5:
      return <Tag color='cyan'>{t('支持人员')}</Tag>;
    case 10:
      return <Tag color='yellow'>{t('管理员')}</Tag>;
    case 100:
      return <Tag color='orange'>{t('超级管理员')}</Tag>;
    default:
      return <Tag color='grey'>{t('未知身份')}</Tag>;
  }
};

const renderUserStatus = (status, t) => {
  if (status === 1) {
    return <Tag color='green'>{t('已启用')}</Tag>;
  }
  if (status === 2) {
    return <Tag color='red'>{t('已禁用')}</Tag>;
  }
  return <Tag color='grey'>{t('未知状态')}</Tag>;
};

const renderTokenStatus = (status, deleted, t) => {
  if (deleted) {
    return <Tag color='red'>{t('已删除')}</Tag>;
  }
  if (status === 1) {
    return <Tag color='green'>{t('可用')}</Tag>;
  }
  if (status === 2) {
    return <Tag color='red'>{t('已禁用')}</Tag>;
  }
  if (status === 3) {
    return <Tag color='orange'>{t('已耗尽')}</Tag>;
  }
  if (status === 4) {
    return <Tag color='grey'>{t('已过期')}</Tag>;
  }
  return <Tag color='grey'>{t('未知状态')}</Tag>;
};

const formatTimestamp = (timestamp, t) => {
  if (!timestamp || timestamp <= 0) {
    return '-';
  }
  if (timestamp === -1) {
    return t('永不过期');
  }
  return timestamp2string(timestamp);
};

const UserApiKeySearch = () => {
  const { t } = useTranslation();
  const [searchKey, setSearchKey] = useState('');
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);
  const [result, setResult] = useState(null);

  const handleSearch = async () => {
    const currentKey = searchKey.trim();
    if (!currentKey) {
      showError(t('请输入 API Key'));
      return;
    }

    setLoading(true);
    setSearched(true);

    try {
      const res = await API.get('/api/user/search/api-key', {
        params: {
          key: currentKey,
        },
      });
      const { success, message, data } = res.data;
      if (!success) {
        setResult(null);
        showError(message);
        return;
      }
      setResult(data);
    } catch (error) {
      setResult(null);
      showError(t('查询失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  const renderContent = () => {
    if (!searched) {
      return (
        <Empty
          image={<UserSearch size={56} color='var(--semi-color-text-2)' />}
          title={t('输入 API Key 后开始查询')}
          description={t('支持直接粘贴带或不带 sk- 前缀的 Key')}
        />
      );
    }

    if (!result?.user || !result?.token) {
      return (
        <Empty
          image={<UserSearch size={56} color='var(--semi-color-text-2)' />}
          title={t('未找到匹配用户')}
          description={t('请确认 API Key 是否完整且仍然存在')}
        />
      );
    }

    const { user, token } = result;

    return (
      <div className='grid grid-cols-1 xl:grid-cols-2 gap-4'>
        <Card
          bordered={false}
          className='!rounded-2xl shadow-sm border border-[var(--semi-color-border)]'
          title={
            <div className='flex items-center gap-2'>
              <UserSearch size={18} />
              <span>{t('用户信息')}</span>
            </div>
          }
        >
          <Descriptions>
            <Descriptions.Item itemKey={t('用户 ID')}>
              {user.id}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('CAH ID')}>
              {user.cah_id || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('用户名')}>
              {user.username || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('显示名称')}>
              {user.display_name || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('邮箱')}>
              {user.email || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('分组')}>
              {user.group || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('角色')}>
              {renderUserRole(user.role, t)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('状态')}>
              {renderUserStatus(user.status, t)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('剩余额度')}>
              {renderQuota(user.quota || 0)}
            </Descriptions.Item>
          </Descriptions>
        </Card>

        <Card
          bordered={false}
          className='!rounded-2xl shadow-sm border border-[var(--semi-color-border)]'
          title={
            <div className='flex items-center gap-2'>
              <KeyRound size={18} />
              <span>{t('令牌信息')}</span>
            </div>
          }
        >
          <Descriptions>
            <Descriptions.Item itemKey={t('令牌 ID')}>
              {token.id}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('所属用户 ID')}>
              {token.user_id}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('令牌名称')}>
              {token.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('分组')}>
              {token.group || '-'}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('状态')}>
              {renderTokenStatus(token.status, token.deleted, t)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('创建时间')}>
              {formatTimestamp(token.created_time, t)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('最后使用')}>
              {formatTimestamp(token.accessed_time, t)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('过期时间')}>
              {formatTimestamp(token.expired_time, t)}
            </Descriptions.Item>
          </Descriptions>
        </Card>
      </div>
    );
  };

  return (
    <div className='mt-[60px] px-2'>
      <CardPro
        type='type1'
        descriptionArea={
          <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
            <div className='flex items-center text-blue-500'>
              <KeyRound size={18} className='mr-2' />
              <Text>{t('API Key 查用户')}</Text>
            </div>
            <Text type='secondary' className='text-sm'>
              {t('输入完整 API Key，系统会自动兼容带或不带 sk- 前缀的内容')}
            </Text>
          </div>
        }
        actionsArea={
          <div className='flex flex-col md:flex-row gap-2 w-full'>
            <Input
              value={searchKey}
              onChange={setSearchKey}
              placeholder={t('请输入 API Key，例如 sk-xxx 或 xxx')}
              prefix={<IconSearch />}
              showClear
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  handleSearch();
                }
              }}
            />
            <Space>
              <Button theme='solid' loading={loading} onClick={handleSearch}>
                {t('查询')}
              </Button>
              <Button
                type='tertiary'
                onClick={() => {
                  setSearchKey('');
                  setSearched(false);
                  setResult(null);
                }}
              >
                {t('清空')}
              </Button>
            </Space>
          </div>
        }
        t={t}
      >
        {renderContent()}
      </CardPro>
    </div>
  );
};

export default UserApiKeySearch;
