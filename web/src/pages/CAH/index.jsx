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

import React, { useContext } from 'react';
import { Button, Card, Tag, Typography } from '@douyinfe/semi-ui';
import { Copy, Fingerprint } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import { copy, getCurrentUser, showError, showSuccess } from '../../helpers';

const CAH = () => {
  const { t } = useTranslation();
  const [userState] = useContext(UserContext);
  const currentUser = userState?.user || getCurrentUser();
  const cahId = currentUser?.cah_id || '';
  const userLabel =
    currentUser?.display_name || currentUser?.username || t('当前账户');

  const handleCopy = async () => {
    if (!cahId) {
      showError(t('未获取到 CAH，请刷新页面后重试'));
      return;
    }

    const ok = await copy(cahId);
    if (ok) {
      showSuccess(t('CAH 已复制'));
    } else {
      showError(t('复制失败，请手动复制'));
    }
  };

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto max-w-3xl'>
        <Card className='!rounded-2xl border-0 shadow-sm'>
          <div className='flex flex-col gap-6'>
            <div className='flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between'>
              <div className='flex items-start gap-3'>
                <div className='flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full bg-semi-color-primary-light-default text-semi-color-primary'>
                  <Fingerprint size={22} strokeWidth={2.2} />
                </div>
                <div className='min-w-0'>
                  <Typography.Title heading={4} className='!mb-1'>
                    {t('当前登录账户 CAH')}
                  </Typography.Title>
                  <Typography.Text type='secondary' className='break-words'>
                    {t('账户')}：{userLabel}
                  </Typography.Text>
                </div>
              </div>
              <Tag color='blue' size='large' shape='circle'>
                CAH
              </Tag>
            </div>

            <div className='rounded-2xl border border-semi-color-border bg-semi-color-fill-0 p-4 sm:p-6'>
              <Typography.Text strong type='secondary' className='!mb-3 !block'>
                {t('当前登录账户的 CAH')}
              </Typography.Text>
              <div className='flex flex-col gap-3 md:flex-row md:items-center md:justify-between'>
                <div className='min-w-0 font-mono text-3xl font-bold tracking-wide text-semi-color-text-0 sm:text-4xl'>
                  {cahId || '-'}
                </div>
                <Button
                  type='primary'
                  theme='solid'
                  icon={<Copy size={16} />}
                  disabled={!cahId}
                  onClick={handleCopy}
                  className='self-start md:self-auto'
                >
                  {t('复制 CAH')}
                </Button>
              </div>
            </div>

            <Typography.Text type='tertiary'>
              {t('需要提供账户标识时，请复制这串 CAH。')}
            </Typography.Text>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default CAH;
