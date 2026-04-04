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

import HeaderBar from './headerbar';
import { Banner, Button, Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar';
import App from '../../App';
import FooterBar from './Footer';
import { ToastContainer } from 'react-toastify';
import React, { useContext, useEffect, useState } from 'react';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { useSidebarCollapsed } from '../../hooks/common/useSidebarCollapsed';
import { useTranslation } from 'react-i18next';
import {
  API,
  getLogo,
  getSystemName,
  setUserData,
  showError,
  showSuccess,
  setStatusData,
  updateAPI,
} from '../../helpers';
import { UserContext } from '../../context/User';
import { StatusContext } from '../../context/Status';
import { useLocation, useNavigate } from 'react-router-dom';
const { Sider, Content, Header } = Layout;

const PageLayout = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [, statusDispatch] = useContext(StatusContext);
  const isMobile = useIsMobile();
  const [collapsed, , setCollapsed] = useSidebarCollapsed();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const { i18n } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();

  const cardProPages = [
    '/console/channel',
    '/console/log',
    '/console/usage',
    '/console/redemption',
    '/console/topup-coupon',
    '/console/user',
    '/console/messages',
    '/console/message-management',
    '/console/token',
    '/console/midjourney',
    '/console/task',
    '/console/models',
    '/pricing',
  ];

  const shouldHideFooter = cardProPages.includes(location.pathname);

  const shouldInnerPadding =
    location.pathname.includes('/console') &&
    !location.pathname.startsWith('/console/chat') &&
    location.pathname !== '/console/playground';

  const isConsoleRoute = location.pathname.startsWith('/console');
  const showSider = isConsoleRoute && (!isMobile || drawerOpen);

  useEffect(() => {
    if (isMobile && drawerOpen && collapsed) {
      setCollapsed(false);
    }
  }, [isMobile, drawerOpen, collapsed, setCollapsed]);

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };

  const refreshUserFromServer = async (skipErrorHandler = false) => {
    const localUser = localStorage.getItem('user');
    if (!localUser) {
      return;
    }

    try {
      const res = await API.get('/api/user/self', {
        disableDuplicate: true,
        skipErrorHandler,
      });
      const { success, data } = res.data;
      if (success && data) {
        userDispatch({ type: 'login', payload: data });
        setUserData(data);
        updateAPI();
      }
    } catch (error) {
      if (!skipErrorHandler) {
        showError(error);
      }
    }
  };

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {
        statusDispatch({ type: 'set', payload: data });
        setStatusData(data);
      } else {
        showError('Unable to connect to server');
      }
    } catch (error) {
      showError('Failed to load status');
    }
  };

  useEffect(() => {
    loadUser();
    loadStatus().catch(console.error);
    refreshUserFromServer(true).catch(() => null);
    let systemName = getSystemName();
    if (systemName) {
      document.title = systemName;
    }
    let logo = getLogo();
    if (logo) {
      let linkElement = document.querySelector("link[rel~='icon']");
      if (linkElement) {
        linkElement.href = logo;
      }
    }
    const savedLang = localStorage.getItem('i18nextLng');
    if (savedLang) {
      i18n.changeLanguage(savedLang);
    }
  }, [i18n]);

  useEffect(() => {
    const user = localStorage.getItem('user');
    if (!user) {
      return undefined;
    }

    const intervalId = window.setInterval(() => {
      refreshUserFromServer(true).catch(() => null);
    }, 60000);

    return () => window.clearInterval(intervalId);
  }, [location.pathname]);

  const handleStopImpersonation = async () => {
    try {
      const res = await API.post('/api/user/impersonation/stop');
      if (res.data.success && res.data.data?.user) {
        userDispatch({ type: 'login', payload: res.data.data.user });
        setUserData(res.data.data.user);
        updateAPI();
        showSuccess('已返回原始管理员会话');
        navigate('/console/user');
      } else {
        showError(res.data.message || '无法返回原始会话');
      }
    } catch (error) {
      showError(error);
    }
  };

  return (
    <Layout
      style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: isMobile ? 'visible' : 'hidden',
      }}
    >
      <Header
        style={{
          padding: 0,
          height: 'auto',
          lineHeight: 'normal',
          position: 'fixed',
          width: '100%',
          top: 0,
          zIndex: 100,
        }}
      >
        <HeaderBar
          onMobileMenuToggle={() => setDrawerOpen((prev) => !prev)}
          drawerOpen={drawerOpen}
        />
      </Header>
      <Layout
        style={{
          overflow: isMobile ? 'visible' : 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {showSider && (
          <Sider
            style={{
              position: 'fixed',
              left: 0,
              top: '64px',
              zIndex: 99,
              border: 'none',
              paddingRight: '0',
              height: 'calc(100vh - 64px)',
              width: 'var(--sidebar-current-width)',
            }}
          >
            <SiderBar
              onNavigate={() => {
                if (isMobile) setDrawerOpen(false);
              }}
            />
          </Sider>
        )}
        <Layout
          style={{
            marginLeft: isMobile
              ? '0'
              : showSider
                ? 'var(--sidebar-current-width)'
                : '0',
            flex: '1 1 auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Content
            style={{
              flex: '1 0 auto',
              overflowY: isMobile ? 'visible' : 'hidden',
              WebkitOverflowScrolling: 'touch',
              padding: shouldInnerPadding ? (isMobile ? '5px' : '24px') : '0',
              position: 'relative',
            }}
          >
            {userState?.user?.impersonation?.active ? (
              <div className='mt-16 px-2'>
                <Banner
                  type='warning'
                  className='!rounded-2xl'
                  title='当前正在仿冒用户会话'
                  description={
                    <div className='flex flex-col gap-3 md:flex-row md:items-center md:justify-between'>
                      <div>
                        {userState.user.impersonation.operator_username ||
                          '管理员'}{' '}
                        正在以当前账户身份访问系统。
                        {userState.user.impersonation.read_only
                          ? ' 当前为只读会话。'
                          : ' 当前为标准会话。'}
                      </div>
                      <Button type='primary' onClick={handleStopImpersonation}>
                        返回原始会话
                      </Button>
                    </div>
                  }
                />
              </div>
            ) : null}

            {!userState?.user?.impersonation?.active &&
            userState?.user?.break_glass_alert ? (
              <div className='mt-16 px-2'>
                <Banner
                  type='danger'
                  className='!rounded-2xl'
                  title='检测到打破玻璃访问记录'
                  description={
                    <div className='flex flex-col gap-3 md:flex-row md:items-center md:justify-between'>
                      <div>
                        {userState.user.break_glass_alert.operator_username ||
                          '客服人员'}{' '}
                        {userState.user.break_glass_alert.active
                          ? '正在通过打破玻璃方式访问你的账户。'
                          : '最近通过打破玻璃方式访问过你的账户。'}
                      </div>
                      <Button
                        theme='solid'
                        type='danger'
                        onClick={() => navigate('/console/personal')}
                      >
                        查看详情
                      </Button>
                    </div>
                  }
                />
              </div>
            ) : null}

            <App />
          </Content>
          {!shouldHideFooter && (
            <Layout.Footer
              style={{
                flex: '0 0 auto',
                width: '100%',
              }}
            >
              <FooterBar />
            </Layout.Footer>
          )}
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  );
};

export default PageLayout;
