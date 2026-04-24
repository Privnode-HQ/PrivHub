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
  InputNumber,
  Select,
  Space,
  Typography,
} from '@douyinfe/semi-ui';
import { Search, Trash2 } from 'lucide-react';
import { API, showError, showSuccess, showWarning } from '../../../helpers';
import { useTranslation } from 'react-i18next';
import {
  createEmptyMultiplierRule,
  formatUserOption,
  mergeUserOptions,
  normalizeMultiplierRulesForSave,
  parseMultiplierRules,
  usageLimitMetricOptions,
  usageLimitTargetOptions,
} from './rateLimitUtils';

const { Paragraph, Text } = Typography;

export default function SettingsUsageLimitMultiplier({ options, refresh }) {
  const { t } = useTranslation();

  const [rules, setRules] = useState([]);
  const [groupOptions, setGroupOptions] = useState([]);
  const [userOptions, setUserOptions] = useState([]);
  const [loadingRuleKey, setLoadingRuleKey] = useState('');
  const [saving, setSaving] = useState(false);

  const targetOptions = useMemo(
    () =>
      usageLimitTargetOptions.map((option) => ({
        label: t(option.label),
        value: option.value,
      })),
    [t],
  );

  const metricOptions = useMemo(
    () =>
      usageLimitMetricOptions.map((option) => ({
        label: option.label,
        value: option.value,
      })),
    [],
  );

  const savedRules = useMemo(
    () => parseMultiplierRules(options.UserUsageLimitMultiplierRules || '[]'),
    [options.UserUsageLimitMultiplierRules],
  );

  const savedValue = useMemo(
    () => JSON.stringify(normalizeMultiplierRulesForSave(savedRules)),
    [savedRules],
  );
  const currentValue = JSON.stringify(normalizeMultiplierRulesForSave(rules));
  const hasChanges = savedValue !== currentValue;

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

  const hydrateUsersByIds = async (userIds) => {
    const uniqueIds = [...new Set(userIds.filter(Boolean))].filter(
      (userId) => !userOptions.some((option) => option.value === userId),
    );
    if (uniqueIds.length === 0) {
      return;
    }

    try {
      const results = await Promise.all(
        uniqueIds.map(async (userId) => {
          try {
            const res = await API.get(`/api/user/${userId}`);
            if (!res.data.success) {
              return null;
            }
            return formatUserOption(res.data.data);
          } catch {
            return null;
          }
        }),
      );

      setUserOptions((prev) =>
        mergeUserOptions(
          prev,
          results.filter(Boolean),
        ),
      );
    } catch (error) {
      showError(error.message);
    }
  };

  useEffect(() => {
    void loadGroups();
  }, []);

  useEffect(() => {
    setRules(savedRules);
    void hydrateUsersByIds(
      savedRules.flatMap((rule) => (Array.isArray(rule.user_ids) ? rule.user_ids : [])),
    );
  }, [savedRules]);

  const updateRule = (localKey, updater) => {
    setRules((prev) =>
      prev.map((rule) =>
        rule.localKey === localKey
          ? {
              ...rule,
              ...(typeof updater === 'function' ? updater(rule) : updater),
            }
          : rule,
      ),
    );
  };

  const removeRule = (localKey) => {
    setRules((prev) => prev.filter((rule) => rule.localKey !== localKey));
  };

  const searchUsers = async (localKey, keyword) => {
    const trimmedKeyword = `${keyword || ''}`.trim();
    if (!trimmedKeyword) {
      showWarning(t('请输入用户名、邮箱或 ID 搜索用户'));
      return;
    }

    setLoadingRuleKey(localKey);
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
      setLoadingRuleKey('');
    }
  };

  const validateRules = () => {
    for (let index = 0; index < rules.length; index += 1) {
      const rule = rules[index];
      const ruleNumber = index + 1;

      if (!rule.metrics || rule.metrics.length === 0) {
        showWarning(t(`请为规则 ${ruleNumber} 选择至少一个限制指标`));
        return false;
      }
      if (rule.scope === 'groups' && (!rule.group_names || rule.group_names.length === 0)) {
        showWarning(t(`请为规则 ${ruleNumber} 选择至少一个分组`));
        return false;
      }
      if (rule.scope === 'users' && (!rule.user_ids || rule.user_ids.length === 0)) {
        showWarning(t(`请为规则 ${ruleNumber} 选择至少一个用户`));
        return false;
      }
      if (!Number.isFinite(Number(rule.multiplier)) || Number(rule.multiplier) <= 0) {
        showWarning(t(`规则 ${ruleNumber} 的倍率必须大于 0`));
        return false;
      }
    }
    return true;
  };

  const saveRules = async () => {
    if (!validateRules()) {
      return;
    }

    setSaving(true);
    try {
      const value = JSON.stringify(normalizeMultiplierRulesForSave(rules));
      const res = await API.put('/api/option/', {
        key: 'UserUsageLimitMultiplierRules',
        value,
      });

      if (!res.data.success) {
        showError(res.data.message);
        return;
      }

      showSuccess(t('保存成功'));
      refresh();
    } catch (error) {
      showError(error.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Space vertical align='start' style={{ width: '100%' }} spacing='medium'>
      <div>
        <Text strong>{t('使用限制倍率')}</Text>
        <Paragraph type='secondary' style={{ marginTop: 8, marginBottom: 0 }}>
          {t(
            '倍率会基于基础使用限制放大或缩小指定指标。优先级按“全部（非封禁）用户 < 特定分组 < 特定用户”覆盖，同层规则按列表顺序后者覆盖前者。',
          )}
        </Paragraph>
        <Paragraph type='secondary' style={{ marginTop: 8, marginBottom: 0 }}>
          {t(
            '倍率只作用于已配置基础使用限制的指标，不会为原本未配置的指标创建新限制。',
          )}
        </Paragraph>
      </div>

      {rules.map((rule, index) => (
        <div
          key={rule.localKey}
          style={{
            width: '100%',
            border: '1px solid var(--semi-color-border)',
            borderRadius: 12,
            padding: 16,
          }}
        >
          <Space vertical align='start' style={{ width: '100%' }} spacing='medium'>
            <Space
              align='center'
              style={{
                width: '100%',
                justifyContent: 'space-between',
              }}
            >
              <Text strong>{`${t('规则')} ${index + 1}`}</Text>
              <Button
                theme='borderless'
                type='danger'
                icon={<Trash2 size={16} />}
                onClick={() => removeRule(rule.localKey)}
              />
            </Space>

            <Select
              value={rule.scope}
              optionList={targetOptions}
              style={{ width: '100%', maxWidth: 360 }}
              onChange={(value) =>
                updateRule(rule.localKey, {
                  scope: value,
                  group_names: value === 'groups' ? rule.group_names : [],
                  user_ids: value === 'users' ? rule.user_ids : [],
                })
              }
            />

            {rule.scope === 'groups' && (
              <Select
                multiple
                value={rule.group_names}
                optionList={groupOptions}
                placeholder={t('请选择一个或多个分组')}
                style={{ width: '100%' }}
                onChange={(value) =>
                  updateRule(rule.localKey, { group_names: value || [] })
                }
              />
            )}

            {rule.scope === 'users' && (
              <Space vertical align='start' style={{ width: '100%' }}>
                <Space style={{ width: '100%' }}>
                  <Input
                    value={rule.userSearchKeyword}
                    prefix={<Search size={14} />}
                    placeholder={t('输入用户名、邮箱或 ID 搜索用户')}
                    onChange={(value) =>
                      updateRule(rule.localKey, { userSearchKeyword: value })
                    }
                    onEnterPress={() =>
                      searchUsers(rule.localKey, rule.userSearchKeyword)
                    }
                  />
                  <Button
                    loading={loadingRuleKey === rule.localKey}
                    onClick={() => searchUsers(rule.localKey, rule.userSearchKeyword)}
                  >
                    {t('搜索')}
                  </Button>
                </Space>
                <Select
                  multiple
                  value={rule.user_ids}
                  optionList={userOptions}
                  placeholder={t('请选择一个或多个用户')}
                  style={{ width: '100%' }}
                  onChange={(value) =>
                    updateRule(rule.localKey, { user_ids: value || [] })
                  }
                />
              </Space>
            )}

            <Select
              multiple
              value={rule.metrics}
              optionList={metricOptions}
              placeholder={t('请选择一个或多个限制指标')}
              style={{ width: '100%' }}
              onChange={(value) =>
                updateRule(rule.localKey, { metrics: value || [] })
              }
            />

            <InputNumber
              value={rule.multiplier}
              min={0.01}
              precision={4}
              step={0.05}
              style={{ width: '100%', maxWidth: 240 }}
              onChange={(value) =>
                updateRule(rule.localKey, { multiplier: Number(value) || 0 })
              }
            />
          </Space>
        </div>
      ))}

      <Space>
        <Button theme='light' onClick={() => setRules((prev) => [...prev, createEmptyMultiplierRule()])}>
          {t('新增倍率规则')}
        </Button>
        <Button loading={saving} disabled={!hasChanges} onClick={saveRules}>
          {t('保存倍率规则')}
        </Button>
      </Space>
    </Space>
  );
}
