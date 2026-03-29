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

import React, { useContext, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Space,
  Switch,
  Tabs,
  TabPane,
  Typography,
} from '@douyinfe/semi-ui';
import { Bell, Settings } from 'lucide-react';
import { API, showError, showSuccess } from '../../../../helpers';
import { StatusContext } from '../../../../context/Status';
import { useSidebar } from '../../../../hooks/common/useSidebar';
import { useUserPermissions } from '../../../../hooks/common/useUserPermissions';

const { Text } = Typography;

const buildDefaultSidebarModules = () => ({
  chat: {
    enabled: true,
    playground: true,
    chat: true,
  },
  console: {
    enabled: true,
    detail: true,
    token: true,
    log: true,
    usage: true,
    midjourney: true,
    task: true,
  },
  personal: {
    enabled: true,
    topup: true,
    personal: true,
    message: true,
    support: true,
  },
  admin: {
    enabled: true,
    channel: true,
    message_manage: true,
    models: true,
    redemption: true,
    topup_coupon: true,
    user: true,
    setting: true,
  },
});

const mergeSidebarModules = (config) => {
  const defaults = buildDefaultSidebarModules();
  const merged = {};

  Object.keys(defaults).forEach((sectionKey) => {
    merged[sectionKey] = {
      ...defaults[sectionKey],
      ...(config?.[sectionKey] || {}),
    };
  });

  return merged;
};

const NotificationSettings = ({
  t,
  notificationSettings,
  handleNotificationSettingChange,
  saveNotificationSettings,
}) => {
  const [statusState] = useContext(StatusContext);
  const [activeTabKey, setActiveTabKey] = useState('notification');
  const [sidebarLoading, setSidebarLoading] = useState(false);
  const [sidebarModulesUser, setSidebarModulesUser] = useState(
    buildDefaultSidebarModules(),
  );

  const { refreshUserConfig } = useSidebar();
  const {
    hasSidebarSettingsPermission,
    isSidebarSectionAllowed,
    isSidebarModuleAllowed,
  } = useUserPermissions();

  useEffect(() => {
    const loadSidebarConfigs = async () => {
      try {
        const userRes = await API.get('/api/user/self');
        if (userRes.data.success && userRes.data.data.sidebar_modules) {
          const userConf = JSON.parse(userRes.data.data.sidebar_modules);
          setSidebarModulesUser(mergeSidebarModules(userConf));
        } else {
          setSidebarModulesUser(buildDefaultSidebarModules());
        }
      } catch (error) {
        console.error('加载边栏配置失败:', error);
      }
    };

    loadSidebarConfigs();
  }, [statusState]);

  const handleSectionChange = (sectionKey, checked) => {
    setSidebarModulesUser((prev) => ({
      ...prev,
      [sectionKey]: {
        ...prev[sectionKey],
        enabled: checked,
      },
    }));
  };

  const handleModuleChange = (sectionKey, moduleKey, checked) => {
    setSidebarModulesUser((prev) => ({
      ...prev,
      [sectionKey]: {
        ...prev[sectionKey],
        [moduleKey]: checked,
      },
    }));
  };

  const saveSidebarSettings = async () => {
    setSidebarLoading(true);
    try {
      const res = await API.put('/api/user/self', {
        sidebar_modules: JSON.stringify(sidebarModulesUser),
      });
      if (res.data.success) {
        showSuccess(t('侧边栏设置保存成功'));
        await refreshUserConfig();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('保存失败'));
    } finally {
      setSidebarLoading(false);
    }
  };

  const sidebarSections = [
    {
      key: 'chat',
      title: t('聊天'),
      modules: [
        { key: 'playground', label: t('操练场') },
        { key: 'chat', label: t('聊天') },
      ],
    },
    {
      key: 'console',
      title: t('控制台'),
      modules: [
        { key: 'detail', label: t('数据看板') },
        { key: 'token', label: t('令牌管理') },
        { key: 'log', label: t('使用日志') },
        { key: 'usage', label: t('使用限制') },
        { key: 'midjourney', label: t('绘图日志') },
        { key: 'task', label: t('任务日志') },
      ],
    },
    {
      key: 'personal',
      title: t('个人中心'),
      modules: [
        { key: 'topup', label: t('钱包管理') },
        { key: 'personal', label: t('个人设置') },
        { key: 'message', label: t('我的信息') },
        { key: 'support', label: t('联系支持') },
      ],
    },
    {
      key: 'admin',
      title: t('管理员'),
      modules: [
        { key: 'channel', label: t('渠道管理') },
        { key: 'message_manage', label: t('信息管理') },
        { key: 'models', label: t('模型管理') },
        { key: 'redemption', label: t('兑换码管理') },
        { key: 'topup_coupon', label: t('折扣中心') },
        { key: 'user', label: t('用户管理') },
        { key: 'setting', label: t('系统设置') },
      ],
    },
  ].filter((section) => isSidebarSectionAllowed(section.key));

  const renderNotificationTab = () => (
    <Card
      bordered={false}
      headerLine={false}
      title={t('通知配置')}
      headerExtraContent={
        <Button type='primary' onClick={saveNotificationSettings}>
          {t('保存设置')}
        </Button>
      }
    >
      <Space vertical align='start' spacing='tight' style={{ width: '100%' }}>
        <Text type='secondary'>
          {t(
            '额度预警将通过站内消息和邮件发送。若未填写通知邮箱，则默认使用账号绑定邮箱。',
          )}
        </Text>
      </Space>

      <Form style={{ marginTop: 24 }}>
        <Form.InputNumber
          label={t('预警阈值')}
          min={1}
          value={notificationSettings.warningThreshold}
          onChange={(value) =>
            handleNotificationSettingChange('warningThreshold', value)
          }
          extraText={t('当剩余额度低于此数值时，系统将通过站内和邮件提醒您')}
        />

        <Form.Input
          label={t('通知邮箱')}
          value={notificationSettings.notificationEmail}
          placeholder={t('留空则使用账号邮箱')}
          onChange={(value) =>
            handleNotificationSettingChange('notificationEmail', value)
          }
        />

        <Form.Switch
          field='acceptUnsetModelRatioModel'
          label={t('接受未设置价格的模型')}
          checked={notificationSettings.acceptUnsetModelRatioModel}
          onChange={(checked) =>
            handleNotificationSettingChange(
              'acceptUnsetModelRatioModel',
              checked,
            )
          }
        />

        <Form.Switch
          field='recordIpLog'
          label={t('记录请求和错误日志 IP')}
          checked={notificationSettings.recordIpLog}
          onChange={(checked) =>
            handleNotificationSettingChange('recordIpLog', checked)
          }
        />
      </Form>
    </Card>
  );

  const renderSidebarTab = () => (
    <Card
      bordered={false}
      headerLine={false}
      title={t('侧边栏设置')}
      headerExtraContent={
        <Space>
          <Button
            theme='light'
            onClick={() => setSidebarModulesUser(buildDefaultSidebarModules())}
          >
            {t('恢复默认')}
          </Button>
          <Button
            type='primary'
            loading={sidebarLoading}
            onClick={saveSidebarSettings}
          >
            {t('保存设置')}
          </Button>
        </Space>
      }
    >
      <Space vertical align='start' spacing='tight' style={{ width: '100%' }}>
        <Text type='secondary'>
          {t(
            '可在这里控制左侧导航中显示哪些入口。管理员关闭的模块不会出现在此处。',
          )}
        </Text>
      </Space>

      <div className='mt-6 space-y-4'>
        {sidebarSections.map((section) => (
          <Card key={section.key} shadows='hover' style={{ marginTop: 16 }}>
            <div className='flex items-center justify-between gap-3'>
              <div>
                <div className='font-semibold'>{section.title}</div>
                <Text type='secondary'>
                  {t('开启后显示该分组下允许访问的模块')}
                </Text>
              </div>
              <Switch
                checked={sidebarModulesUser[section.key]?.enabled !== false}
                onChange={(checked) =>
                  handleSectionChange(section.key, checked)
                }
              />
            </div>

            <div className='mt-4 grid grid-cols-1 md:grid-cols-2 gap-3'>
              {section.modules
                .filter((module) =>
                  isSidebarModuleAllowed(section.key, module.key),
                )
                .map((module) => (
                  <div
                    key={module.key}
                    className='flex items-center justify-between rounded-xl border border-semi-color-border bg-semi-color-fill-0 px-4 py-3'
                  >
                    <span>{module.label}</span>
                    <Switch
                      checked={
                        sidebarModulesUser[section.key]?.[module.key] !== false
                      }
                      disabled={
                        sidebarModulesUser[section.key]?.enabled === false
                      }
                      onChange={(checked) =>
                        handleModuleChange(section.key, module.key, checked)
                      }
                    />
                  </div>
                ))}
            </div>
          </Card>
        ))}
      </div>
    </Card>
  );

  return (
    <Card style={{ minHeight: '100%' }}>
      <Tabs activeKey={activeTabKey} onChange={setActiveTabKey} type='line'>
        <TabPane
          itemKey='notification'
          tab={
            <span className='flex items-center gap-2'>
              <Bell size={16} />
              {t('通知')}
            </span>
          }
        >
          {renderNotificationTab()}
        </TabPane>

        {hasSidebarSettingsPermission() && (
          <TabPane
            itemKey='sidebar'
            tab={
              <span className='flex items-center gap-2'>
                <Settings size={16} />
                {t('侧边栏')}
              </span>
            }
          >
            {renderSidebarTab()}
          </TabPane>
        )}
      </Tabs>
    </Card>
  );
};

export default NotificationSettings;
