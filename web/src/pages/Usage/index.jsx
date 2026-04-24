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
import { Button, Card, Empty, Spin, Tag, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import { renderQuota } from '../../helpers';
import useUserUsageLimits from '../../hooks/usage/useUserUsageLimits';

const budgetMetricDefs = [
  { key: 'hourly', titleKey: '每小时预算' },
  { key: 'daily', titleKey: '每日预算' },
  { key: 'weekly', titleKey: '每周预算' },
  { key: 'monthly', titleKey: '月度预算' },
];

const tableMetricDefs = [
  { key: 'rpm', titleKey: '每分钟请求数' },
  { key: 'rpd', titleKey: '每日请求数' },
  { key: 'tpm', titleKey: '每分钟 Token' },
  { key: 'tpd', titleKey: '每日 Token' },
];

const formatPlainNumber = (value) =>
  Number(value || 0).toLocaleString(undefined, {
    maximumFractionDigits: 0,
  });

const formatPercentage = (value) => {
  if (value === null || value === undefined) {
    return null;
  }
  return `${Number(value).toFixed(Number(value) % 1 === 0 ? 0 : 1)}%`;
};

const getStatusColor = (status) => {
  switch (status) {
    case 'blocked':
      return 'red';
    case 'unlimited':
      return 'grey';
    default:
      return 'green';
  }
};

const getProgressColor = (percent) => {
  if (percent >= 90) return 'bg-red-500';
  if (percent >= 70) return 'bg-amber-500';
  return 'bg-emerald-500';
};

const getStatusDotColor = (status) => {
  switch (status) {
    case 'blocked':
      return 'bg-red-500';
    case 'unlimited':
      return 'bg-zinc-300';
    default:
      return 'bg-emerald-500';
  }
};

export default function Usage() {
  const { t } = useTranslation();
  const { data, loading, refreshing, error, refresh } = useUserUsageLimits();

  const renderMetricValue = (metric, field) => {
    const value = metric?.[field];
    if (value === null || value === undefined) {
      return t('不限制');
    }
    if (metric.unit === 'quota') {
      return renderQuota(value);
    }
    return formatPlainNumber(value);
  };

  const renderStatus = (metric) => {
    switch (metric.status) {
      case 'blocked':
        return t('已达上限');
      case 'unlimited':
        return t('无限制');
      default:
        return t('可用');
    }
  };

  const renderConsumptionPercent = (metric) => {
    const formatted = formatPercentage(metric?.consumption_percent);
    if (formatted) {
      return formatted;
    }
    if (metric?.status === 'unlimited') {
      return t('不限制');
    }
    return '-';
  };

  const budgetMetrics = budgetMetricDefs
    .map((def) => ({ ...def, metric: data?.metrics?.[def.key] }))
    .filter((m) => m.metric);

  const tableMetrics = tableMetricDefs
    .map((def) => ({ ...def, metric: data?.metrics?.[def.key] }))
    .filter((m) => m.metric);

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto max-w-7xl space-y-4'>
        {/* Header */}
        <Card className='!rounded-2xl border-0 shadow-sm'>
          <div className='flex flex-col gap-4 md:flex-row md:items-start md:justify-between'>
            <div className='space-y-2'>
              <Typography.Title heading={4} className='!mb-0'>
                {t('使用限制')}
              </Typography.Title>
              <Typography.Text type='secondary'>
                {t(
                  '查看当前分组的请求数、Token 以及每小时、每日、每周、月度预算限制。',
                )}
              </Typography.Text>
              {data && (
                <div className='flex flex-wrap gap-2 pt-1'>
                  <Tag color='blue'>
                    {t('当前分组')}：{data.group || t('默认')}
                  </Tag>
                  <Tag color='grey'>
                    {t('当前站点额度单位')}：{data.billing_unit || 'USD'}
                  </Tag>
                  {data.legacy_group_rate_limit_replaced ? (
                    <Tag color='green'>{t('旧分组速率限制已被替换')}</Tag>
                  ) : null}
                </div>
              )}
            </div>
            <Button
              type='primary'
              theme='solid'
              loading={refreshing}
              onClick={refresh}
            >
              {t('刷新')}
            </Button>
          </div>
        </Card>

        <Spin spinning={loading}>
          {!loading && !data ? (
            <Card className='!rounded-2xl border-0 shadow-sm'>
              <Empty description={error || t('暂无可显示的使用限制数据')} />
            </Card>
          ) : null}

          {data ? (
            <>
              {data.no_limits_configured ? (
                <Card className='!rounded-2xl border-0 shadow-sm'>
                  <Typography.Text type='secondary'>
                    {t(
                      '当前分组未配置任何限制，以下数据仅用于说明当前状态。',
                    )}
                  </Typography.Text>
                </Card>
              ) : null}

              {/* Budget Limits — Progress Bars */}
              {budgetMetrics.length > 0 && (
                <Card className='!rounded-2xl border-0 shadow-sm'>
                  <Typography.Text
                    strong
                    type='secondary'
                    className='!mb-4 !block uppercase tracking-wide'
                    size='small'
                  >
                    {t('预算限制')}
                  </Typography.Text>
                  <div className='divide-y divide-zinc-950/5'>
                    {budgetMetrics.map(({ key, titleKey, metric }) => {
                      const percent = metric.consumption_percent ?? 0;
                      const clampedPercent = Math.min(
                        100,
                        Math.max(0, percent),
                      );

                      return (
                        <div
                          key={key}
                          className='space-y-2 py-4 first:pt-0 last:pb-0'
                        >
                          <div className='flex items-center justify-between gap-3'>
                            <div className='flex items-center gap-2'>
                              <Typography.Text strong className='truncate'>
                                {t(titleKey)}
                              </Typography.Text>
                              <Tag
                                size='small'
                                color={getStatusColor(metric.status)}
                              >
                                {renderStatus(metric)}
                              </Tag>
                            </div>
                            <Typography.Text
                              className='tabular-nums'
                              type='secondary'
                            >
                              {metric.hide_details
                                ? renderConsumptionPercent(metric)
                                : `${renderMetricValue(metric, 'used')} / ${renderMetricValue(metric, 'limit')}`}
                            </Typography.Text>
                          </div>
                          {metric.status !== 'unlimited' && (
                            <div className='h-2 overflow-hidden rounded-full bg-zinc-100'>
                              <div
                                className={`h-full rounded-full transition-all ${getProgressColor(clampedPercent)}`}
                                style={{ width: `${clampedPercent}%` }}
                              />
                            </div>
                          )}
                          {!metric.hide_details && (
                            <div className='flex items-center justify-between gap-3'>
                              <Typography.Text type='tertiary' size='small'>
                                {t('剩余')}:{' '}
                                {renderMetricValue(metric, 'remaining')}
                                {metric.pending
                                  ? ` · ${t('处理中')}: ${renderMetricValue(metric, 'pending')}`
                                  : ''}
                              </Typography.Text>
                              <Typography.Text
                                type='tertiary'
                                size='small'
                                className='tabular-nums'
                              >
                                {metric.reset_at
                                  ? `${t('重置时间')}: ${dayjs(metric.reset_at).format('MM-DD HH:mm')}`
                                  : ''}
                              </Typography.Text>
                            </div>
                          )}
                          {metric.hide_details && metric.reset_at && (
                            <div className='flex justify-end'>
                              <Typography.Text
                                type='tertiary'
                                size='small'
                                className='tabular-nums'
                              >
                                {t('重置时间')}:{' '}
                                {dayjs(metric.reset_at).format('MM-DD HH:mm')}
                              </Typography.Text>
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </Card>
              )}

              {/* Request & Token Limits — Compact Table */}
              {tableMetrics.length > 0 && (
                <Card className='!rounded-2xl border-0 shadow-sm overflow-hidden'>
                  <Typography.Text
                    strong
                    type='secondary'
                    className='!mb-4 !block uppercase tracking-wide'
                    size='small'
                  >
                    {t('请求与 Token 限制')}
                  </Typography.Text>
                  <div className='overflow-x-auto'>
                    <table className='w-full'>
                      <thead>
                        <tr className='border-b border-zinc-950/5'>
                          <th className='px-4 py-3 text-left'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('指标')}
                            </Typography.Text>
                          </th>
                          <th className='px-4 py-3 text-center'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('状态')}
                            </Typography.Text>
                          </th>
                          <th className='px-4 py-3 text-right'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('已用')} / {t('限制')}
                            </Typography.Text>
                          </th>
                          <th className='hidden px-4 py-3 text-right sm:table-cell'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('剩余')}
                            </Typography.Text>
                          </th>
                          <th className='px-4 py-3 text-right'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('消耗')}
                            </Typography.Text>
                          </th>
                          <th className='hidden px-4 py-3 text-right md:table-cell'>
                            <Typography.Text
                              type='secondary'
                              size='small'
                              strong
                            >
                              {t('重置时间')}
                            </Typography.Text>
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {tableMetrics.map(({ key, titleKey, metric }, index) => {
                          const percent = metric.consumption_percent ?? 0;
                          const clampedPercent = Math.min(
                            100,
                            Math.max(0, percent),
                          );

                          return (
                            <tr
                              key={key}
                              className={
                                index < tableMetrics.length - 1
                                  ? 'border-b border-zinc-950/5'
                                  : ''
                              }
                            >
                              <td className='px-4 py-3'>
                                <Typography.Text strong>
                                  {t(titleKey)}
                                </Typography.Text>
                              </td>
                              <td className='px-4 py-3 text-center'>
                                <span className='inline-flex items-center gap-1.5'>
                                  <span
                                    className={`size-2 rounded-full ${getStatusDotColor(metric.status)}`}
                                  />
                                  <Typography.Text size='small'>
                                    {renderStatus(metric)}
                                  </Typography.Text>
                                </span>
                              </td>
                              <td className='px-4 py-3 text-right'>
                                <Typography.Text className='tabular-nums'>
                                  {metric.hide_details
                                    ? '-'
                                    : `${renderMetricValue(metric, 'used')} / ${renderMetricValue(metric, 'limit')}`}
                                </Typography.Text>
                              </td>
                              <td className='hidden px-4 py-3 text-right sm:table-cell'>
                                <Typography.Text className='tabular-nums'>
                                  {metric.hide_details
                                    ? '-'
                                    : renderMetricValue(metric, 'remaining')}
                                </Typography.Text>
                              </td>
                              <td className='px-4 py-3 text-right'>
                                <div className='flex items-center justify-end gap-2'>
                                  {metric.status !== 'unlimited' && (
                                    <div className='hidden h-1.5 w-16 overflow-hidden rounded-full bg-zinc-100 sm:block'>
                                      <div
                                        className={`h-full rounded-full ${getProgressColor(clampedPercent)}`}
                                        style={{
                                          width: `${clampedPercent}%`,
                                        }}
                                      />
                                    </div>
                                  )}
                                  <Typography.Text
                                    className='tabular-nums'
                                    size='small'
                                  >
                                    {renderConsumptionPercent(metric)}
                                  </Typography.Text>
                                </div>
                              </td>
                              <td className='hidden px-4 py-3 text-right md:table-cell'>
                                <Typography.Text
                                  className='tabular-nums'
                                  type='secondary'
                                  size='small'
                                >
                                  {metric.reset_at
                                    ? dayjs(metric.reset_at).format(
                                        'MM-DD HH:mm',
                                      )
                                    : '-'}
                                </Typography.Text>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </Card>
              )}
            </>
          ) : null}
        </Spin>
      </div>
    </div>
  );
}
