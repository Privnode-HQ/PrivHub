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
  Table,
  Typography,
} from '@douyinfe/semi-ui';
import { Eye, EyeOff, Plus, Trash2 } from 'lucide-react';
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

export default function RequestRateLimit(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [groups, setGroups] = useState([]);
  const [savedJson, setSavedJson] = useState('{}');
  const [newGroupName, setNewGroupName] = useState('');

  useEffect(() => {
    const raw = props.options.UserGroupUsageLimits || '{}';
    setSavedJson(raw);
    setGroups(parseGroupLimits(raw));
  }, [props.options.UserGroupUsageLimits]);

  const hasChanges = useMemo(
    () => serializeGroupLimits(groups) !== serializeGroupLimits(parseGroupLimits(savedJson)),
    [groups, savedJson],
  );

  const updateGroup = useCallback((groupName, field, value) => {
    setGroups((prev) =>
      prev.map((g) => (g.groupName === groupName ? { ...g, [field]: value } : g)),
    );
  }, []);

  const renameGroup = useCallback((oldName, newName) => {
    setGroups((prev) =>
      prev.map((g) => (g.groupName === oldName ? { ...g, groupName: newName } : g)),
    );
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
  }, []);

  const removeGroup = useCallback((groupName) => {
    setGroups((prev) => prev.filter((g) => g.groupName !== groupName));
  }, []);

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

  const metricRows = ALL_METRICS.map((m) => ({
    key: m,
    metric: METRIC_LABELS[m],
    description: t(METRIC_DESCRIPTIONS[m]),
  }));

  const columns = [
    {
      title: t('指标'),
      dataIndex: 'metric',
      fixed: 'left',
      width: 150,
      render: (text, record) => (
        <div>
          <Text strong size='small'>{text}</Text>
          <br />
          <Text type='tertiary' size='small'>{record.description}</Text>
        </div>
      ),
    },
    ...groups.map((group) => ({
      title: (
        <Space vertical spacing={2} align='center'>
          <Input
            size='small'
            value={group.groupName}
            style={{ width: 120, textAlign: 'center' }}
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
              icon={<Trash2 size={12} />}
              style={{ padding: '1px 3px' }}
            />
          </Popconfirm>
        </Space>
      ),
      dataIndex: group.groupName,
      width: 160,
      render: (_, record) => (
        <Space vertical spacing={2} align='start' style={{ width: '100%' }}>
          <InputNumber
            value={group[record.key]}
            min={0}
            step={1}
            size='small'
            placeholder='∞'
            style={{ width: '100%' }}
            onChange={(v) =>
              updateGroup(
                group.groupName,
                record.key,
                v === '' || v === undefined || v === null ? null : Number(v),
              )
            }
          />
          <Space align='center' spacing={4}>
            <Switch
              size='small'
              checked={group[`${record.key}_hide_details`]}
              onChange={(v) => updateGroup(group.groupName, `${record.key}_hide_details`, v)}
            />
            <Text type='tertiary' size='small'>
              {group[`${record.key}_hide_details`] ? <EyeOff size={12} /> : <Eye size={12} />}
            </Text>
          </Space>
        </Space>
      ),
    })),
  ];

  return (
    <Spin spinning={loading}>
      <Space vertical align='start' style={{ width: '100%' }} spacing='medium'>
        <div>
          <Text strong size='normal'>{t('基础使用限制')}</Text>
          <Paragraph type='secondary' size='small' style={{ marginTop: 4, marginBottom: 0 }}>
            {t('纵向按指标对比各分组配置。留空表示不限制。')}
          </Paragraph>
          <Paragraph type='secondary' size='small' style={{ marginTop: 4, marginBottom: 0 }}>
            {t('开启隐藏详情后，用户侧仅显示消耗百分比，不显示已用、处理中、剩余和总限制。')}
          </Paragraph>
        </div>

        <div style={{ width: '100%', overflow: 'auto' }}>
          <Table
            columns={columns}
            dataSource={metricRows}
            rowKey='key'
            pagination={false}
            size='small'
            scroll={{ x: Math.max(150 + groups.length * 160, 500) }}
            empty={
              <Paragraph type='tertiary' style={{ padding: '24px 0' }}>
                {t('暂无分组，点击下方按钮添加')}
              </Paragraph>
            }
          />
        </div>

        <Space>
          <Input
            size='small'
            value={newGroupName}
            placeholder={t('输入分组名称')}
            style={{ width: 160 }}
            onChange={setNewGroupName}
            onEnterPress={handleAddGroup}
          />
          <Button
            icon={<Plus size={14} />}
            theme='light'
            onClick={handleAddGroup}
          >
            {t('添加分组')}
          </Button>
          <Button
            loading={loading}
            disabled={!hasChanges}
            onClick={save}
          >
            {t('保存基础使用限制')}
          </Button>
        </Space>
      </Space>
    </Spin>
  );
}
