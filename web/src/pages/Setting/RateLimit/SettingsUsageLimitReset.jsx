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

import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Input,
  Modal,
  Select,
  Space,
  Typography,
} from '@douyinfe/semi-ui';
import { Search } from 'lucide-react';
import { API, showError, showSuccess, showWarning } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import {
  formatUserOption,
  mergeUserOptions,
  usageLimitTargetOptions,
} from './rateLimitUtils';

const { Paragraph, Text } = Typography;

export default function SettingsUsageLimitReset() {
  const { t } = useTranslation();

  const [scope, setScope] = useState('all');
  const [groupOptions, setGroupOptions] = useState([]);
  const [selectedGroups, setSelectedGroups] = useState([]);
  const [selectedUserIds, setSelectedUserIds] = useState([]);
  const [userOptions, setUserOptions] = useState([]);
  const [userSearchKeyword, setUserSearchKeyword] = useState('');
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [lastResult, setLastResult] = useState(null);

  const scopeOptions = useMemo(
    () =>
      usageLimitTargetOptions.map((option) => ({
        label: t(option.label),
        value: option.value,
      })),
    [t],
  );

  const loadGroups = async () => {
    try {
      const res = await API.get('/api/group/');
      if (!res.data.success) {
        showError(res.data.message);
        return;
      }

      const options = (res.data.data || []).map((groupName) => ({
        label: groupName,
        value: groupName,
      }));
      setGroupOptions(options);
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    void loadGroups();
  }, []);

  const searchUsers = async (keyword = userSearchKeyword) => {
    const trimmedKeyword = `${keyword || ''}`.trim();
    if (!trimmedKeyword) {
      showWarning(t('请输入用户名、邮箱或 ID 搜索用户'));
      return;
    }

    setUserSearchLoading(true);
    try {
      const res = await API.get('/api/user/search', {
        params: {
          keyword: trimmedKeyword,
          p: 1,
          page_size: 20,
        },
      });

      if (!res.data.success) {
        showError(res.data.message);
        return;
      }

      setUserOptions((prev) =>
        mergeUserOptions(
          prev,
          (res.data.data?.items || []).map((user) => formatUserOption(user)),
        ),
      );
    } catch (error) {
      showError(error.message);
    } finally {
      setUserSearchLoading(false);
    }
  };

  const resetUsageLimits = async () => {
    if (scope === 'groups' && selectedGroups.length === 0) {
      showWarning(t('请选择至少一个分组'));
      return;
    }
    if (scope === 'users' && selectedUserIds.length === 0) {
      showWarning(t('请选择至少一个用户'));
      return;
    }

    Modal.confirm({
      title: t('确认重置使用限制'),
      content: t(
        '该操作会清空目标用户当前所有使用限制窗口和未结算预留，且无法撤销。',
      ),
      okText: t('确认重置'),
      cancelText: t('取消'),
      onOk: async () => {
        setSubmitting(true);
        try {
          const res = await API.post('/api/usage/reset', {
            scope,
            group_names: selectedGroups,
            user_ids: selectedUserIds,
          });

          if (!res.data.success) {
            showError(res.data.message);
            return;
          }

          setLastResult(res.data.data || null);
          showSuccess(t('使用限制已重置'));
        } catch (error) {
          showError(error.message);
        } finally {
          setSubmitting(false);
        }
      },
    });
  };

  return (
    <Space vertical align='start' style={{ width: '100%' }} spacing='medium'>
      <div>
        <Text strong>{t('重置使用限制')}</Text>
        <Paragraph type='secondary' style={{ marginTop: 8, marginBottom: 0 }}>
          {t(
            '单独执行一次使用限制重置，不会修改基础使用限制和倍率规则。',
          )}
        </Paragraph>
      </div>

      <Select
        value={scope}
        optionList={scopeOptions}
        style={{ width: '100%', maxWidth: 360 }}
        onChange={(value) => setScope(value)}
      />

      {scope === 'groups' && (
        <Select
          multiple
          value={selectedGroups}
          optionList={groupOptions}
          placeholder={t('请选择一个或多个分组')}
          style={{ width: '100%' }}
          onChange={(value) => setSelectedGroups(value || [])}
        />
      )}

      {scope === 'users' && (
        <Space vertical align='start' style={{ width: '100%' }}>
          <Space style={{ width: '100%' }}>
            <Input
              value={userSearchKeyword}
              prefix={<Search size={14} />}
              placeholder={t('输入用户名、邮箱或 ID 搜索用户')}
              onChange={setUserSearchKeyword}
              onEnterPress={() => searchUsers(userSearchKeyword)}
            />
            <Button
              loading={userSearchLoading}
              onClick={() => searchUsers(userSearchKeyword)}
            >
              {t('搜索')}
            </Button>
          </Space>
          <Select
            multiple
            value={selectedUserIds}
            optionList={userOptions}
            placeholder={t('请选择一个或多个用户')}
            style={{ width: '100%' }}
            onChange={(value) => setSelectedUserIds(value || [])}
          />
        </Space>
      )}

      <Button loading={submitting} theme='solid' onClick={resetUsageLimits}>
        {t('重置使用限制')}
      </Button>

      {lastResult && (
        <Paragraph type='secondary' style={{ marginBottom: 0 }}>
          {t('最近一次执行')}:
          {` ${t('命中用户')} ${lastResult.targeted_users || 0}，${t(
            '清理窗口',
          )} ${lastResult.reset_windows || 0}，${t('清理预留')} ${
            lastResult.reset_reservations || 0
          }`}
          {lastResult.scope === 'all' &&
            `，${t('跳过封禁用户')} ${lastResult.skipped_banned_users || 0}`}
        </Paragraph>
      )}
    </Space>
  );
}
