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

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Input,
  InputNumber,
  Popconfirm,
  Space,
  Spin,
  Switch,
  Typography,
} from '@douyinfe/semi-ui';
import { Eye, EyeOff, Plus, Search, Trash2 } from 'lucide-react';
import {
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

const { Text, Paragraph } = Typography;

const RATE_METRICS = ['rpm', 'rpd', 'tpm', 'tpd'];
const BUDGET_METRICS = ['hourly', 'daily', 'weekly', 'monthly'];
const ALL_METRICS = [...RATE_METRICS, ...BUDGET_METRICS];

const METRIC_LABELS = {
  rpm: 'RPM',
  rpd: 'RPD',
  tpm: 'TPM',
  tpd: 'TPD',
  hourly: 'Hourly',
  daily: 'Daily',
  weekly: 'Weekly',
  monthly: 'Monthly',
};

const METRIC_DESCRIPTIONS = {
  rpm: '请求/分钟',
  rpd: '请求/天',
  tpm: 'Token/分钟',
  tpd: 'Token/天',
  hourly: '每小时预算',
  daily: '每日预算',
  weekly: '每周预算',
  monthly: '每月预算',
};

function parseGroupLimits(jsonString) {
  try {
    const obj = JSON.parse(jsonString || '{}');
    return Object.entries(obj).map(([groupName, limits]) => ({
      groupName,
      ...ALL_METRICS.reduce((acc, m) => {
        acc[m] = limits[m] ?? null;
        acc[`${m}_hide_details`] = !!limits[`${m}_hide_details`];
        return acc;
      }, {}),
    }));
  } catch {
    return [];
  }
}

function serializeGroupLimits(groups) {
  const obj = {};
  groups.forEach((g) => {
    const limits = {};
    ALL_METRICS.forEach((m) => {
      limits[m] = g[m];
      if (g[`${m}_hide_details`]) {
        limits[`${m}_hide_details`] = true;
      }
    });
    obj[g.groupName] = limits;
  });
  return JSON.stringify(obj);
}

function countConfiguredLimits(group) {
  return ALL_METRICS.filter((m) => group[m] != null).length;
}

function MetricRow({ metric, group, updateGroup, t }) {
  return (
    <div className='flex items-center justify-between gap-3 py-2.5 px-1'>
      <div className='min-w-0 flex-shrink-0' style={{ width: 140 }}>
        <Text strong size='small'>{METRIC_LABELS[metric]}</Text>
        <br />
        <Text type='tertiary' size='small'>{t(METRIC_DESCRIPTIONS[metric])}</Text>
      </div>
      <div className='flex items-center gap-3 flex-1 justify-end'>
        <InputNumber
          value={group[metric]}
          min={0}
          step={1}
          size='small'
          placeholder='∞'
          style={{ width: 140 }}
          onChange={(v) =>
            updateGroup(
              group.groupName,
              metric,
              v === '' || v === undefined || v === null ? null : Number(v),
            )
          }
        />
        <Space align='center' spacing={4}>
          <Switch
            size='small'
            checked={group[`${metric}_hide_details`]}
            onChange={(v) => updateGroup(group.groupName, `${metric}_hide_details`, v)}
          />
          <Text type='tertiary' size='small'>
            {group[`${metric}_hide_details`] ? <EyeOff size={12} /> : <Eye size={12} />}
          </Text>
        </Space>
      </div>
    </div>
  );
}

function GroupDetailPanel({ group, updateGroup, renameGroup, removeGroup, t }) {
  if (!group) {
    return (
      <div className='flex items-center justify-center h-full'>
        <Paragraph type='tertiary'>{t('选择一个分组进行编辑')}</Paragraph>
      </div>
    );
  }

  return (
    <div className='p-4'>
      <div className='flex items-center gap-2 mb-5'>
        <Input
          size='default'
          value={group.groupName}
          style={{ width: 200 }}
          onChange={(v) => renameGroup(group.groupName, v)}
        />
        <Popconfirm
          title={t('确认删除该分组？')}
          onConfirm={() => removeGroup(group.groupName)}
        >
          <Button
            size='small'
            theme='borderless'
            type='danger'
            icon={<Trash2 size={14} />}
          />
        </Popconfirm>
      </div>

      <div className='mb-5'>
        <Text strong size='small' className='mb-2 block' style={{ color: 'var(--semi-color-text-2)' }}>
          {t('速率限制')}
        </Text>
        <div className='rounded-lg' style={{ border: '1px solid var(--semi-color-border)' }}>
          {RATE_METRICS.map((m, i) => (
            <div
              key={m}
              style={i < RATE_METRICS.length - 1 ? { borderBottom: '1px solid var(--semi-color-border)' } : undefined}
            >
              <MetricRow metric={m} group={group} updateGroup={updateGroup} t={t} />
            </div>
          ))}
        </div>
      </div>

      <div>
        <Text strong size='small' className='mb-2 block' style={{ color: 'var(--semi-color-text-2)' }}>
          {t('预算限制')}
        </Text>
        <div className='rounded-lg' style={{ border: '1px solid var(--semi-color-border)' }}>
          {BUDGET_METRICS.map((m, i) => (
            <div
              key={m}
              style={i < BUDGET_METRICS.length - 1 ? { borderBottom: '1px solid var(--semi-color-border)' } : undefined}
            >
              <MetricRow metric={m} group={group} updateGroup={updateGroup} t={t} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default function RequestRateLimit(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [groups, setGroups] = useState([]);
  const [savedJson, setSavedJson] = useState('{}');
  const [newGroupName, setNewGroupName] = useState('');
  const [selectedGroupName, setSelectedGroupName] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    const raw = props.options.UserGroupUsageLimits || '{}';
    setSavedJson(raw);
    const parsed = parseGroupLimits(raw);
    setGroups(parsed);
    if (parsed.length > 0 && !selectedGroupName) {
      setSelectedGroupName(parsed[0].groupName);
    }
  }, [props.options.UserGroupUsageLimits]);

  const hasChanges = useMemo(
    () => serializeGroupLimits(groups) !== serializeGroupLimits(parseGroupLimits(savedJson)),
    [groups, savedJson],
  );

  const selectedGroup = useMemo(
    () => groups.find((g) => g.groupName === selectedGroupName) || null,
    [groups, selectedGroupName],
  );

  const filteredGroups = useMemo(() => {
    if (!searchQuery.trim()) return groups;
    const q = searchQuery.trim().toLowerCase();
    return groups.filter((g) => g.groupName.toLowerCase().includes(q));
  }, [groups, searchQuery]);

  const updateGroup = useCallback((groupName, field, value) => {
    setGroups((prev) =>
      prev.map((g) => (g.groupName === groupName ? { ...g, [field]: value } : g)),
    );
  }, []);

  const renameGroup = useCallback((oldName, newName) => {
    setGroups((prev) =>
      prev.map((g) => (g.groupName === oldName ? { ...g, groupName: newName } : g)),
    );
    setSelectedGroupName((prev) => (prev === oldName ? newName : prev));
  }, []);

  const addGroup = useCallback((name = '') => {
    const groupName = name || `group_${Date.now()}`;
    setGroups((prev) => [
      ...prev,
      {
        groupName,
        ...ALL_METRICS.reduce((acc, m) => {
          acc[m] = null;
          acc[`${m}_hide_details`] = false;
          return acc;
        }, {}),
      },
    ]);
    setSelectedGroupName(groupName);
  }, []);

  const removeGroup = useCallback((groupName) => {
    setGroups((prev) => {
      const next = prev.filter((g) => g.groupName !== groupName);
      if (selectedGroupName === groupName) {
        setSelectedGroupName(next.length > 0 ? next[0].groupName : null);
      }
      return next;
    });
  }, [selectedGroupName]);

  const save = useCallback(async () => {
    const names = groups.map((g) => g.groupName);
    if (new Set(names).size !== names.length) {
      showWarning(t('存在重复的分组名称'));
      return;
    }
    if (names.some((n) => !n.trim())) {
      showWarning(t('分组名称不能为空'));
      return;
    }

    const value = serializeGroupLimits(groups);
    setLoading(true);
    try {
      const res = await API.put('/api/option/', {
        key: 'UserGroupUsageLimits',
        value,
      });
      if (!res.data.success) {
        showError(res.data.message);
        return;
      }
      showSuccess(t('保存成功'));
      props.refresh();
    } catch {
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  }, [groups, props.refresh, t]);

  const handleAddGroup = () => {
    const name = newGroupName.trim();
    if (!name) {
      addGroup();
    } else {
      if (groups.some((g) => g.groupName === name)) {
        showWarning(t('分组名称已存在'));
        return;
      }
      addGroup(name);
      setNewGroupName('');
    }
  };

  return (
    <Spin spinning={loading}>
      <Space vertical align='start' style={{ width: '100%' }} spacing='medium'>
        <div>
          <Text strong size='normal'>{t('基础使用限制')}</Text>
          <Paragraph type='secondary' size='small' style={{ marginTop: 4, marginBottom: 0 }}>
            {t('为每个分组配置速率和预算限制。留空表示不限制。')}
          </Paragraph>
          <Paragraph type='secondary' size='small' style={{ marginTop: 4, marginBottom: 0 }}>
            {t('开启隐藏详情后，用户侧仅显示消耗百分比，不显示已用、处理中、剩余和总限制。')}
          </Paragraph>
        </div>

        <div
          className='flex w-full rounded-lg overflow-hidden'
          style={{
            border: '1px solid var(--semi-color-border)',
            minHeight: 420,
          }}
        >
          {/* Left panel — Group list */}
          <div
            className='flex flex-col flex-shrink-0'
            style={{
              width: 240,
              borderRight: '1px solid var(--semi-color-border)',
              background: 'var(--semi-color-fill-0)',
            }}
          >
            {/* Search */}
            <div className='p-2' style={{ borderBottom: '1px solid var(--semi-color-border)' }}>
              <Input
                size='small'
                prefix={<Search size={14} />}
                placeholder={t('搜索分组...')}
                value={searchQuery}
                onChange={setSearchQuery}
                showClear
              />
            </div>

            {/* Group count */}
            <div className='px-3 py-1.5' style={{ borderBottom: '1px solid var(--semi-color-border)' }}>
              <Text type='tertiary' size='small'>
                {searchQuery.trim()
                  ? t('{{filtered}} / {{total}} 个分组', { filtered: filteredGroups.length, total: groups.length })
                  : t('{{total}} 个分组', { total: groups.length })
                }
              </Text>
            </div>

            {/* Scrollable group list */}
            <div className='flex-1 overflow-y-auto'>
              {filteredGroups.length === 0 && groups.length > 0 && (
                <div className='p-3 text-center'>
                  <Text type='tertiary' size='small'>{t('无匹配分组')}</Text>
                </div>
              )}
              {filteredGroups.map((group) => {
                const isSelected = group.groupName === selectedGroupName;
                const configured = countConfiguredLimits(group);
                return (
                  <div
                    key={group.groupName}
                    className='px-3 py-2 cursor-pointer flex items-center justify-between'
                    style={{
                      background: isSelected ? 'var(--semi-color-primary-light-default)' : undefined,
                      borderLeft: isSelected ? '3px solid var(--semi-color-primary)' : '3px solid transparent',
                    }}
                    onClick={() => setSelectedGroupName(group.groupName)}
                  >
                    <Text
                      size='small'
                      strong={isSelected}
                      ellipsis={{ showTooltip: true }}
                      style={{ flex: 1, minWidth: 0 }}
                    >
                      {group.groupName}
                    </Text>
                    <Text type='tertiary' size='small' style={{ flexShrink: 0, marginLeft: 8 }}>
                      {configured}/{ALL_METRICS.length}
                    </Text>
                  </div>
                );
              })}
            </div>

            {/* Add group */}
            <div className='p-2 flex gap-1' style={{ borderTop: '1px solid var(--semi-color-border)' }}>
              <Input
                size='small'
                value={newGroupName}
                placeholder={t('分组名称')}
                className='flex-1'
                onChange={setNewGroupName}
                onEnterPress={handleAddGroup}
              />
              <Button
                size='small'
                icon={<Plus size={14} />}
                theme='light'
                onClick={handleAddGroup}
              />
            </div>
          </div>

          {/* Right panel — Group detail */}
          <div className='flex-1 overflow-y-auto' style={{ background: 'var(--semi-color-bg-0)' }}>
            {groups.length === 0 ? (
              <div className='flex items-center justify-center h-full'>
                <Paragraph type='tertiary' style={{ padding: '24px 0' }}>
                  {t('暂无分组，在左侧添加')}
                </Paragraph>
              </div>
            ) : (
              <GroupDetailPanel
                group={selectedGroup}
                updateGroup={updateGroup}
                renameGroup={renameGroup}
                removeGroup={removeGroup}
                t={t}
              />
            )}
          </div>
        </div>

        <Button
          loading={loading}
          disabled={!hasChanges}
          onClick={save}
        >
          {t('保存基础使用限制')}
        </Button>
      </Space>
    </Spin>
  );
}
