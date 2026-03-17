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

const metricDefinitions = [
  { key: 'rpm', titleKey: '每分钟请求数' },
  { key: 'rpd', titleKey: '每日请求数' },
  { key: 'tpm', titleKey: '每分钟 Token' },
  { key: 'tpd', titleKey: '每日 Token' },
  { key: 'monthly', titleKey: '月度预算' },
];

const formatPlainNumber = (value) =>
  Number(value || 0).toLocaleString(undefined, {
    maximumFractionDigits: 0,
  });

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

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto max-w-7xl space-y-4'>
        <Card className='!rounded-2xl border-0 shadow-sm'>
          <div className='flex flex-col gap-4 md:flex-row md:items-start md:justify-between'>
            <div className='space-y-2'>
              <Typography.Title heading={4} className='!mb-0'>
                {t('使用限制')}
              </Typography.Title>
              <Typography.Text type='secondary'>
                {t('查看当前分组的请求数、Token 和月度预算限制。')}
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
                    {t('当前分组未配置任何限制，以下数据仅用于说明当前状态。')}
                  </Typography.Text>
                </Card>
              ) : null}

              <div className='grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-5'>
                {metricDefinitions.map((definition) => {
                  const metric = data.metrics?.[definition.key];
                  if (!metric) {
                    return null;
                  }

                  return (
                    <Card
                      key={definition.key}
                      className='!rounded-2xl border-0 shadow-sm'
                    >
                      <div className='space-y-4'>
                        <div className='flex items-start justify-between gap-3'>
                          <div className='space-y-1'>
                            <Typography.Text strong>
                              {t(definition.titleKey)}
                            </Typography.Text>
                            <div>
                              <Tag color={getStatusColor(metric.status)}>
                                {renderStatus(metric)}
                              </Tag>
                            </div>
                          </div>
                        </div>

                        <div className='space-y-2'>
                          <div className='flex items-center justify-between gap-3'>
                            <Typography.Text type='secondary'>
                              {t('限制')}
                            </Typography.Text>
                            <Typography.Text strong>
                              {renderMetricValue(metric, 'limit')}
                            </Typography.Text>
                          </div>

                          <div className='flex items-center justify-between gap-3'>
                            <Typography.Text type='secondary'>
                              {t('已用')}
                            </Typography.Text>
                            <Typography.Text>
                              {renderMetricValue(metric, 'used')}
                            </Typography.Text>
                          </div>

                          <div className='flex items-center justify-between gap-3'>
                            <Typography.Text type='secondary'>
                              {t('处理中')}
                            </Typography.Text>
                            <Typography.Text>
                              {renderMetricValue(metric, 'pending')}
                            </Typography.Text>
                          </div>

                          <div className='flex items-center justify-between gap-3'>
                            <Typography.Text type='secondary'>
                              {t('剩余')}
                            </Typography.Text>
                            <Typography.Text>
                              {renderMetricValue(metric, 'remaining')}
                            </Typography.Text>
                          </div>

                          <div className='flex items-start justify-between gap-3'>
                            <Typography.Text type='secondary'>
                              {t('重置时间')}
                            </Typography.Text>
                            <Typography.Text className='text-right'>
                              {metric.reset_at
                                ? dayjs(metric.reset_at).format(
                                    'YYYY-MM-DD HH:mm:ss',
                                  )
                                : '-'}
                            </Typography.Text>
                          </div>
                        </div>
                      </div>
                    </Card>
                  );
                })}
              </div>
            </>
          ) : null}
        </Spin>
      </div>
    </div>
  );
}
