import React from 'react';
import {
  Banner,
  Button,
  Card,
  Empty,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { ShieldAlert, ShieldCheck } from 'lucide-react';
import { formatDateTimeString } from '../../../../helpers';

const { Text } = Typography;

const SupportAccessCard = ({
  t,
  supportAccessState,
  supportAccessLoading,
  onOpenSupportAccess,
  onCloseSupportAccess,
}) => {
  const openSupportAccess = supportAccessState?.open_support_access || null;
  const incidents = supportAccessState?.break_glass_incidents || [];

  return (
    <Card
      className='!rounded-2xl mt-4 md:mt-6'
      title={t('客服访问与打破玻璃')}
      headerExtraContent={
        openSupportAccess ? (
          <Button
            theme='light'
            type='danger'
            loading={supportAccessLoading}
            onClick={onCloseSupportAccess}
          >
            {t('关闭未使用访问')}
          </Button>
        ) : (
          <Button
            type='primary'
            loading={supportAccessLoading}
            onClick={onOpenSupportAccess}
          >
            {t('开放一次客服访问')}
          </Button>
        )
      }
    >
      <Space vertical align='start' spacing='tight' style={{ width: '100%' }}>
        <Text type='secondary'>
          {t(
            '你可以主动开放一次客服访问权限。该权限只能激活一次，管理员必须在 24 小时内使用，激活后的会话会在 24 小时后失效。',
          )}
        </Text>
      </Space>

      {openSupportAccess ? (
        <Banner
          type='success'
          className='mt-4 !rounded-2xl'
          title={t('当前存在一条未使用的客服访问授权')}
          description={
            <div className='flex flex-col gap-2'>
              <div>
                {t('失效时间')}:{' '}
                {openSupportAccess.granted_expires_at
                  ? formatDateTimeString(
                      new Date(openSupportAccess.granted_expires_at),
                    )
                  : '-'}
              </div>
              <div>{t('管理员可在此时间前激活一次访问。')}</div>
            </div>
          }
        />
      ) : null}

      <div className='mt-6'>
        <div className='flex items-center gap-2'>
          <ShieldAlert size={18} />
          <Text className='text-base font-medium'>{t('打破玻璃记录')}</Text>
        </div>
        <div className='mt-1 text-sm text-semi-color-text-2'>
          {t('这里会显示客服通过打破玻璃方式访问你的账户时执行过的操作。')}
        </div>
      </div>

      {incidents.length === 0 ? (
        <Empty
          image={<ShieldCheck size={40} />}
          description={t('暂无打破玻璃记录')}
          style={{ padding: 24 }}
        />
      ) : (
        <div className='mt-4 flex flex-col gap-4'>
          {incidents.map((incident) => (
            <div
              key={incident.id}
              className='rounded-2xl border border-semi-color-border bg-semi-color-bg-1 p-4'
            >
              <div className='flex flex-col gap-2 md:flex-row md:items-center md:justify-between'>
                <div className='flex flex-wrap items-center gap-2'>
                  <Text strong>{incident.operator_username || t('客服')}</Text>
                  {incident.active ? (
                    <Tag color='red'>{t('访问进行中')}</Tag>
                  ) : (
                    <Tag color='orange'>{t('已结束')}</Tag>
                  )}
                </div>
                <div className='text-xs text-semi-color-text-2'>
                  {t('开始')}:{' '}
                  {incident.started_at
                    ? formatDateTimeString(new Date(incident.started_at))
                    : '-'}
                  {' · '}
                  {t('结束')}:{' '}
                  {incident.ended_at
                    ? formatDateTimeString(new Date(incident.ended_at))
                    : t('进行中')}
                </div>
              </div>

              <div className='mt-3 flex flex-wrap gap-2'>
                <Tag color='white'>
                  {t('记录操作数')}: {incident.action_count || 0}
                </Tag>
              </div>

              <div className='mt-4 flex flex-col gap-2'>
                {(incident.actions || []).length === 0 ? (
                  <Text type='secondary'>{t('暂无记录到具体操作。')}</Text>
                ) : (
                  incident.actions.map((action) => (
                    <div
                      key={action.id}
                      className='flex flex-col gap-1 rounded-xl bg-semi-color-bg-0 p-3'
                    >
                      <div className='flex flex-wrap items-center gap-2'>
                        <Tag color='white'>{action.method}</Tag>
                        <Text strong>{action.route || action.path}</Text>
                        <Tag color={action.success ? 'green' : 'red'}>
                          HTTP {action.status_code}
                        </Tag>
                      </div>
                      <div className='text-xs text-semi-color-text-2'>
                        {action.created_at
                          ? formatDateTimeString(new Date(action.created_at))
                          : '-'}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
};

export default SupportAccessCard;
