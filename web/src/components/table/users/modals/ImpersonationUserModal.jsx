import React, { useEffect, useState } from 'react';
import {
  Banner,
  Button,
  Modal,
  Space,
  Switch,
  Typography,
} from '@douyinfe/semi-ui';

const { Text } = Typography;

const ImpersonationUserModal = ({
  visible,
  user,
  loading,
  onCancel,
  onSubmit,
  t,
}) => {
  const [readOnly, setReadOnly] = useState(false);

  useEffect(() => {
    if (visible) {
      setReadOnly(false);
    }
  }, [visible]);

  const handleStandardSubmit = () => {
    onSubmit?.({
      readOnly,
      breakGlass: false,
    });
  };

  const handleBreakGlassSubmit = () => {
    onSubmit?.({
      readOnly: false,
      breakGlass: true,
    });
  };

  return (
    <Modal
      title={t('仿冒用户会话')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={560}
    >
      <div className='flex flex-col gap-4'>
        <div>
          <Text>
            {t('目标用户')}: <Text strong>{user?.username || '-'}</Text>
          </Text>
        </div>

        <Banner
          type='info'
          title={t('标准仿冒')}
          description={t(
            '默认会先向用户发送站内信和邮件。用户点击链接并确认后，管理员将获得一次性访问权限，需在 24 小时内激活，激活后的会话 24 小时后失效。',
          )}
        />

        <div className='rounded-2xl border border-semi-color-border p-4'>
          <div className='flex items-center justify-between gap-3'>
            <div>
              <Text strong>{t('只读会话')}</Text>
              <div className='text-sm text-semi-color-text-2'>
                {t(
                  '启用后只能查看用户数据，不能代表用户执行操作。该说明会明确展示给用户。',
                )}
              </div>
            </div>
            <Switch checked={readOnly} onChange={setReadOnly} />
          </div>
        </div>

        <Banner
          type='danger'
          title={t('打破玻璃')}
          description={t(
            '会立即强制进入用户账户，不受一次性访问或 24 小时限制。系统会向用户发送站内信和邮件，并在用户界面显示横幅与操作记录。',
          )}
        />

        <Space style={{ justifyContent: 'flex-end', width: '100%' }}>
          <Button onClick={onCancel}>{t('取消')}</Button>
          <Button
            loading={loading}
            onClick={handleStandardSubmit}
            type='primary'
          >
            {t('发送访问请求')}
          </Button>
          <Button
            loading={loading}
            theme='solid'
            type='danger'
            onClick={handleBreakGlassSubmit}
          >
            {t('立即打破玻璃')}
          </Button>
        </Space>
      </div>
    </Modal>
  );
};

export default ImpersonationUserModal;
