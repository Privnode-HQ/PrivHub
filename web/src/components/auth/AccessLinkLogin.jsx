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

import React, { useContext, useEffect, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Button, Card } from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { useTranslation } from 'react-i18next';
import { UserContext } from '../../context/User';
import {
  API,
  getLogo,
  getSystemName,
  setUserData,
  showError,
  showSuccess,
  updateAPI,
} from '../../helpers';

const AccessLinkLogin = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();
  const [, userDispatch] = useContext(UserContext);
  const [consuming, setConsuming] = useState(true);
  const [errorMessage, setErrorMessage] = useState('');
  const hasStartedRef = useRef(false);

  const logo = getLogo();
  const systemName = getSystemName();

  useEffect(() => {
    if (hasStartedRef.current) {
      return;
    }
    hasStartedRef.current = true;

    const token = (searchParams.get('token') || '').trim();
    if (!token) {
      setConsuming(false);
      setErrorMessage(t('访问链接无效或不存在'));
      return;
    }

    const consumeLink = async () => {
      try {
        const res = await API.post(
          `/api/user/access_link/${encodeURIComponent(token)}/consume`,
        );
        const { success, message, data } = res.data;
        if (!success || !data?.user) {
          setErrorMessage(message || t('访问链接无效或不存在'));
          showError(message || t('访问链接无效或不存在'));
          return;
        }
        userDispatch({ type: 'login', payload: data.user });
        setUserData(data.user);
        updateAPI();
        showSuccess(t('访问链接已生效'));
        navigate('/console', { replace: true });
      } catch (error) {
        const message =
          error?.response?.data?.message || t('访问链接无效或不存在');
        setErrorMessage(message);
        showError(message);
      } finally {
        setConsuming(false);
      }
    };

    consumeLink();
  }, [navigate, searchParams, t, userDispatch]);

  return (
    <div className='min-h-screen bg-gradient-to-b from-white to-slate-100 px-4 py-10'>
      <div className='mx-auto flex min-h-[70vh] max-w-xl items-center justify-center'>
        <Card className='w-full !rounded-2xl border-0 shadow-sm'>
          <div className='flex flex-col gap-5 px-2 py-6 text-center'>
            <div className='flex items-center justify-center gap-3'>
              <img src={logo} alt='Logo' className='h-10 rounded-full' />
              <Title heading={3} className='!mb-0'>
                {systemName}
              </Title>
            </div>

            <div className='flex flex-col gap-2'>
              <Title heading={4} className='!mb-0'>
                {consuming ? t('正在验证访问链接') : t('账户访问链接')}
              </Title>
              <Text type='secondary'>
                {consuming
                  ? t('请稍候，系统正在为您创建登录会话。')
                  : errorMessage || t('访问链接已处理完成。')}
              </Text>
            </div>

            {!consuming ? (
              <div className='flex justify-center'>
                <Button type='primary' onClick={() => navigate('/login')}>
                  {t('返回登录页')}
                </Button>
              </div>
            ) : null}
          </div>
        </Card>
      </div>
    </div>
  );
};

export default AccessLinkLogin;
