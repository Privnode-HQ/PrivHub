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
import {
  API,
  copy,
  formatDateTimeString,
  showError,
  showInfo,
  showSuccess,
  setStatusData,
  prepareCredentialCreationOptions,
  buildRegistrationResult,
  isPasskeySupported,
  setUserData,
} from '../../helpers';
import { UserContext } from '../../context/User';
import {
  Banner,
  Button,
  Card,
  Input,
  Modal,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

// 导入子组件
import UserInfoHeader from './personal/components/UserInfoHeader';
import AccountManagement from './personal/cards/AccountManagement';
import NotificationSettings from './personal/cards/NotificationSettings';
import SupportAccessCard from './personal/cards/SupportAccessCard';
import EmailBindModal from './personal/modals/EmailBindModal';
import WeChatBindModal from './personal/modals/WeChatBindModal';
import AccountDeleteModal from './personal/modals/AccountDeleteModal';
import ChangePasswordModal from './personal/modals/ChangePasswordModal';
import SecureVerificationModal from '../common/modals/SecureVerificationModal';
import { useSecureVerification } from '../../hooks/common/useSecureVerification';

const PersonalSetting = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const { t } = useTranslation();
  const { Text } = Typography;

  const [inputs, setInputs] = useState({
    display_name: '',
    wechat_verification_code: '',
    email_verification_code: '',
    email: '',
    self_account_deletion_confirmation: '',
    original_password: '',
    set_new_password: '',
    set_new_password_confirmation: '',
  });
  const [status, setStatus] = useState({});
  const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
  const [showWeChatBindModal, setShowWeChatBindModal] = useState(false);
  const [showEmailBindModal, setShowEmailBindModal] = useState(false);
  const [showAccountDeleteModal, setShowAccountDeleteModal] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const [systemToken, setSystemToken] = useState('');
  const [passkeyStatus, setPasskeyStatus] = useState({ enabled: false });
  const [passkeyRegisterLoading, setPasskeyRegisterLoading] = useState(false);
  const [passkeyDeleteLoading, setPasskeyDeleteLoading] = useState(false);
  const [passkeySupported, setPasskeySupported] = useState(false);
  const [backToPayAsYouGoLoading, setBackToPayAsYouGoLoading] = useState(false);
  const [profileSaving, setProfileSaving] = useState(false);
  const [supportAccessState, setSupportAccessState] = useState({
    break_glass_incidents: [],
  });
  const [supportAccessLoading, setSupportAccessLoading] = useState(false);
  const [approvalRequest, setApprovalRequest] = useState(null);
  const [approvalRequestLoading, setApprovalRequestLoading] = useState(false);
  const [notificationSettings, setNotificationSettings] = useState({
    warningThreshold: 100000,
    notificationEmail: '',
    acceptUnsetModelRatioModel: false,
    recordIpLog: false,
  });
  const requiredActionPromptKeyRef = useRef('');
  const requiredActions = userState?.user?.required_actions || [];
  const requiredActionKey = [...requiredActions].sort().join('|');
  const requiresDisplayName = requiredActions.includes('update_display_name');
  const requiresEmailBind = requiredActions.includes('bind_email');
  const requiresPasswordChange = requiredActions.includes('change_password');
  const hasRequiredActions = requiredActions.length > 0;
  const {
    isModalVisible: isSecureVerificationVisible,
    verificationMethods,
    verificationState,
    executeVerification,
    cancelVerification,
    setVerificationCode,
    switchVerificationMethod,
    withVerification,
  } = useSecureVerification();

  useEffect(() => {
    let saved = localStorage.getItem('status');
    if (saved) {
      const parsed = JSON.parse(saved);
      setStatus(parsed);
      if (parsed.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(parsed.turnstile_site_key);
      } else {
        setTurnstileEnabled(false);
        setTurnstileSiteKey('');
      }
    }
    // Always refresh status from server to avoid stale flags (e.g., admin just enabled OAuth)
    (async () => {
      try {
        const res = await API.get('/api/status');
        const { success, data } = res.data;
        if (success && data) {
          setStatus(data);
          setStatusData(data);
          if (data.turnstile_check) {
            setTurnstileEnabled(true);
            setTurnstileSiteKey(data.turnstile_site_key);
          } else {
            setTurnstileEnabled(false);
            setTurnstileSiteKey('');
          }
        }
      } catch (e) {
        // ignore and keep local status
      }
    })();

    getUserData();
    loadImpersonationHistory();

    isPasskeySupported()
      .then(setPasskeySupported)
      .catch(() => setPasskeySupported(false));
  }, []);

  useEffect(() => {
    const token = searchParams.get('support_access_token');
    if (!token) {
      setApprovalRequest(null);
      return;
    }
    loadApprovalRequest(token);
  }, [searchParams]);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval); // Clean up on unmount
  }, [disableButton, countdown]);

  useEffect(() => {
    if (userState?.user?.setting) {
      const settings = JSON.parse(userState.user.setting);
      setNotificationSettings({
        warningThreshold: settings.quota_warning_threshold || 500000,
        notificationEmail: settings.notification_email || '',
        acceptUnsetModelRatioModel:
          settings.accept_unset_model_ratio_model || false,
        recordIpLog: settings.record_ip_log || false,
      });
    }
  }, [userState?.user?.setting]);

  useEffect(() => {
    if (!requiredActionKey) {
      requiredActionPromptKeyRef.current = '';
      return;
    }
    if (requiredActionPromptKeyRef.current === requiredActionKey) {
      return;
    }
    if (requiresPasswordChange) {
      setShowChangePasswordModal(true);
    } else if (requiresEmailBind) {
      setShowEmailBindModal(true);
    }
    requiredActionPromptKeyRef.current = requiredActionKey;
  }, [requiredActionKey, requiresEmailBind, requiresPasswordChange]);

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const generateAccessToken = async () => {
    const res = await API.get('/api/user/token');
    const { success, message, data } = res.data;
    if (success) {
      setSystemToken(data);
      await copy(data);
      showSuccess(t('令牌已重置并已复制到剪贴板'));
    } else {
      showError(message);
    }
  };

  const loadPasskeyStatus = async () => {
    try {
      const res = await API.get('/api/user/passkey');
      const { success, data, message } = res.data;
      if (success) {
        setPasskeyStatus({
          enabled: data?.enabled || false,
          last_used_at: data?.last_used_at || null,
          backup_eligible: data?.backup_eligible || false,
          backup_state: data?.backup_state || false,
        });
      } else {
        showError(message);
      }
    } catch (error) {
      // 忽略错误，保留默认状态
    }
  };

  const handleRegisterPasskey = async () => {
    if (!passkeySupported || !window.PublicKeyCredential) {
      showInfo(t('当前设备不支持 Passkey'));
      return;
    }
    setPasskeyRegisterLoading(true);
    try {
      const beginRes = await API.post('/api/user/passkey/register/begin');
      const { success, message, data } = beginRes.data;
      if (!success) {
        showError(message || t('无法发起 Passkey 注册'));
        return;
      }

      const publicKey = prepareCredentialCreationOptions(
        data?.options || data?.publicKey || data,
      );
      const credential = await navigator.credentials.create({ publicKey });
      const payload = buildRegistrationResult(credential);
      if (!payload) {
        showError(t('Passkey 注册失败，请重试'));
        return;
      }

      const finishRes = await API.post(
        '/api/user/passkey/register/finish',
        payload,
      );
      if (finishRes.data.success) {
        showSuccess(t('Passkey 注册成功'));
        await loadPasskeyStatus();
      } else {
        showError(finishRes.data.message || t('Passkey 注册失败，请重试'));
      }
    } catch (error) {
      if (error?.name === 'AbortError') {
        showInfo(t('已取消 Passkey 注册'));
      } else {
        showError(t('Passkey 注册失败，请重试'));
      }
    } finally {
      setPasskeyRegisterLoading(false);
    }
  };

  const handleRemovePasskey = async () => {
    setPasskeyDeleteLoading(true);
    try {
      const res = await API.delete('/api/user/passkey');
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('Passkey 已解绑'));
        await loadPasskeyStatus();
      } else {
        showError(message || t('操作失败，请重试'));
      }
    } catch (error) {
      showError(t('操作失败，请重试'));
    } finally {
      setPasskeyDeleteLoading(false);
    }
  };

  const getUserData = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
      setUserData(data);
      setInputs((prev) => ({
        ...prev,
        display_name: data.display_name || '',
        email: data.email || prev.email || '',
      }));
      await loadPasskeyStatus();
    } else {
      showError(message);
    }
  };

  const loadImpersonationHistory = async () => {
    try {
      const res = await API.get('/api/user/impersonation/history');
      const { success, message, data } = res.data;
      if (success) {
        setSupportAccessState(data || { break_glass_incidents: [] });
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('加载客服访问记录失败'));
    }
  };

  const clearSupportAccessToken = () => {
    if (!searchParams.get('support_access_token')) {
      return;
    }
    const nextSearchParams = new URLSearchParams(searchParams);
    nextSearchParams.delete('support_access_token');
    setSearchParams(nextSearchParams, { replace: true });
  };

  const loadApprovalRequest = async (token) => {
    if (!token) {
      setApprovalRequest(null);
      return;
    }

    try {
      const res = await API.get(
        `/api/user/impersonation/request/${encodeURIComponent(token)}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setApprovalRequest({
          ...data,
          token,
        });
      } else {
        showError(message);
        clearSupportAccessToken();
      }
    } catch (error) {
      showError(t('该访问请求不存在或已失效'));
      clearSupportAccessToken();
    }
  };

  const handleApproveAccessRequest = async () => {
    if (!approvalRequest?.token) {
      return;
    }
    try {
      setApprovalRequestLoading(true);
      const approveRequestApiCall = async () => {
        const res = await API.post(
          `/api/user/impersonation/request/${encodeURIComponent(
            approvalRequest.token,
          )}/approve`,
        );
        if (!res.data.success) {
          throw new Error(res.data.message || t('操作失败，请重试'));
        }
        showSuccess(res.data.message || t('访问请求已批准'));
        setApprovalRequest((prev) =>
          prev
            ? {
                ...prev,
                state: 'approved',
              }
            : prev,
        );
        clearSupportAccessToken();
        await loadImpersonationHistory();
        await getUserData();
        return res;
      };
      const result = await withVerification(approveRequestApiCall, {
        title: t('批准客服访问'),
        description: approvalRequest.requested_read_only
          ? t(
              '该请求为只读会话。若你已启用 2FA 或 Passkey，批准前需要先完成一次安全验证。',
            )
          : t(
              '该请求允许客服代表你执行操作。若你已启用 2FA 或 Passkey，批准前需要先完成一次安全验证。',
            ),
      });

      if (!result) {
        return;
      }
    } finally {
      setApprovalRequestLoading(false);
    }
  };

  const handleRejectAccessRequest = async () => {
    if (!approvalRequest?.token) {
      return;
    }

    try {
      setApprovalRequestLoading(true);
      const res = await API.post(
        `/api/user/impersonation/request/${encodeURIComponent(
          approvalRequest.token,
        )}/reject`,
      );
      if (res.data.success) {
        showSuccess(res.data.message || t('访问请求已拒绝'));
        setApprovalRequest((prev) =>
          prev
            ? {
                ...prev,
                state: 'rejected',
              }
            : prev,
        );
        clearSupportAccessToken();
        await loadImpersonationHistory();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('操作失败，请重试'));
    } finally {
      setApprovalRequestLoading(false);
    }
  };

  const handleOpenSupportAccess = async () => {
    try {
      setSupportAccessLoading(true);
      const openSupportAccessApiCall = async () => {
        const res = await API.post('/api/user/impersonation/open_access');
        if (!res.data.success) {
          throw new Error(res.data.message || t('操作失败，请重试'));
        }
        showSuccess(res.data.message || t('已开放一次客服访问'));
        await loadImpersonationHistory();
        return res;
      };
      const result = await withVerification(openSupportAccessApiCall, {
        title: t('开放一次客服访问'),
        description: t(
          '该权限仅可被管理员激活一次，且必须在 24 小时内开始使用。若你已启用 2FA 或 Passkey，需要先完成一次安全验证。',
        ),
      });

      if (!result) {
        return;
      }
    } finally {
      setSupportAccessLoading(false);
    }
  };

  const handleCloseSupportAccess = async () => {
    try {
      setSupportAccessLoading(true);
      const res = await API.delete('/api/user/impersonation/open_access');
      if (res.data.success) {
        showSuccess(res.data.message || t('未使用的客服访问已关闭'));
        await loadImpersonationHistory();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('操作失败，请重试'));
    } finally {
      setSupportAccessLoading(false);
    }
  };

  const saveProfile = async () => {
    const nextDisplayName = (inputs.display_name || '').trim();
    if (
      !nextDisplayName &&
      (requiresDisplayName || userState?.user?.require_display_name_enabled)
    ) {
      showError(t('请输入用户名称'));
      return;
    }

    setProfileSaving(true);
    try {
      const res = await API.put('/api/user/self', {
        display_name: nextDisplayName,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('用户名称已更新'));
        await getUserData();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(t('更新失败，请重试'));
    } finally {
      setProfileSaving(false);
    }
  };

  const handleSystemTokenClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(t('系统令牌已复制到剪切板'));
  };

  const deleteAccount = async () => {
    if (inputs.self_account_deletion_confirmation !== userState.user.username) {
      showError(t('请输入你的账户名以确认删除！'));
      return;
    }

    const res = await API.delete('/api/user/self');
    const { success, message } = res.data;

    if (success) {
      showSuccess(t('账户已删除！'));
      await API.get('/api/user/logout');
      userDispatch({ type: 'logout' });
      localStorage.removeItem('user');
      navigate('/login');
    } else {
      showError(message);
    }
  };

  const bindWeChat = async () => {
    if (inputs.wechat_verification_code === '') return;
    const res = await API.get(
      `/api/oauth/wechat/bind?code=${inputs.wechat_verification_code}`,
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('微信账户绑定成功！'));
      setShowWeChatBindModal(false);
    } else {
      showError(message);
    }
  };

  const changePassword = async () => {
    if (inputs.original_password === '') {
      showError(t('请输入原密码！'));
      return;
    }
    if (inputs.set_new_password === '') {
      showError(t('请输入新密码！'));
      return;
    }
    if (inputs.original_password === inputs.set_new_password) {
      showError(t('新密码需要和原密码不一致！'));
      return;
    }
    if (inputs.set_new_password !== inputs.set_new_password_confirmation) {
      showError(t('两次输入的密码不一致！'));
      return;
    }
    const res = await API.put(`/api/user/self`, {
      original_password: inputs.original_password,
      password: inputs.set_new_password,
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('密码修改成功！'));
      await getUserData();
    } else {
      showError(message);
    }
    setShowChangePasswordModal(false);
  };

  const sendVerificationCode = async () => {
    if (inputs.email === '') {
      showError(t('请输入邮箱！'));
      return;
    }
    setDisableButton(true);
    if (turnstileEnabled && turnstileToken === '') {
      showInfo(t('请稍后几秒重试，Turnstile 正在检查用户环境！'));
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`,
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('验证码发送成功，请检查邮箱！'));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const bindEmail = async () => {
    if (inputs.email_verification_code === '') {
      showError(t('请输入邮箱验证码！'));
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/oauth/email/bind?email=${inputs.email}&code=${inputs.email_verification_code}`,
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('邮箱账户绑定成功！'));
      setShowEmailBindModal(false);
      await getUserData();
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制：') + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  const handleNotificationSettingChange = (type, value) => {
    setNotificationSettings((prev) => ({
      ...prev,
      [type]: value.target
        ? value.target.value !== undefined
          ? value.target.value
          : value.target.checked
        : value, // handle checkbox properly
    }));
  };

  const saveNotificationSettings = async () => {
    try {
      const res = await API.put('/api/user/setting', {
        notify_type: 'email',
        quota_warning_threshold: parseFloat(
          notificationSettings.warningThreshold,
        ),
        notification_email: notificationSettings.notificationEmail,
        accept_unset_model_ratio_model:
          notificationSettings.acceptUnsetModelRatioModel,
        record_ip_log: notificationSettings.recordIpLog,
      });

      if (res.data.success) {
        showSuccess(t('设置保存成功'));
        await getUserData();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('设置保存失败'));
    }
  };

  const backToPayAsYouGo = async () => {
    if (userState?.user?.group !== 'subscription') {
      showInfo(t('当前账户不是订阅分组，无需操作'));
      return;
    }

    setBackToPayAsYouGoLoading(true);
    try {
      const res = await API.post('/api/user/self/back_to_payg');
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('已切换回按量付费'));
        await getUserData();
      } else {
        showError(message || t('操作失败'));
      }
    } catch (e) {
      showError(t('操作失败'));
    } finally {
      setBackToPayAsYouGoLoading(false);
    }
  };

  return (
    <div className='mt-[60px]'>
      <div className='flex justify-center'>
        <div className='w-full max-w-7xl mx-auto px-2'>
          {/* 顶部用户信息区域 */}
          <UserInfoHeader t={t} userState={userState} />

          {hasRequiredActions ? (
            <Banner
              type='warning'
              className='mt-4 !rounded-2xl'
              description={
                <div className='flex flex-col gap-3'>
                  <div>
                    {t(
                      '管理员或系统要求你先完成以下操作后才能继续使用其他功能。',
                    )}
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    {requiresDisplayName ? (
                      <Tag color='orange'>{t('填写用户名称')}</Tag>
                    ) : null}
                    {requiresEmailBind ? (
                      <Tag color='blue'>{t('绑定邮箱')}</Tag>
                    ) : null}
                    {requiresPasswordChange ? (
                      <Tag color='red'>{t('修改密码')}</Tag>
                    ) : null}
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    {requiresDisplayName ? (
                      <Button
                        theme='solid'
                        type='warning'
                        onClick={saveProfile}
                      >
                        {t('保存用户名称')}
                      </Button>
                    ) : null}
                    {requiresEmailBind ? (
                      <Button
                        theme='outline'
                        type='primary'
                        onClick={() => setShowEmailBindModal(true)}
                      >
                        {t('去绑定邮箱')}
                      </Button>
                    ) : null}
                    {requiresPasswordChange ? (
                      <Button
                        theme='outline'
                        type='danger'
                        onClick={() => setShowChangePasswordModal(true)}
                      >
                        {t('去修改密码')}
                      </Button>
                    ) : null}
                  </div>
                </div>
              }
            />
          ) : null}

          <Card className='!rounded-2xl mt-4 md:mt-6'>
            <div className='flex items-center justify-between gap-3 mb-4'>
              <div>
                <Text className='text-lg font-medium'>{t('个人资料')}</Text>
                <div className='text-xs text-gray-600'>
                  {t('用户名称可由你自行编辑，用于展示身份信息')}
                </div>
              </div>
              {requiresDisplayName ? (
                <Tag color='orange'>{t('当前必须填写')}</Tag>
              ) : null}
            </div>

            <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
              <div>
                <Text strong>{t('登录用户名')}</Text>
                <Input
                  value={userState?.user?.username || ''}
                  disabled
                  className='mt-2 !rounded-lg'
                />
              </div>
              <div>
                <Text strong>{t('用户名称')}</Text>
                <Input
                  value={inputs.display_name}
                  onChange={(value) => handleInputChange('display_name', value)}
                  placeholder={t('请输入用户名称')}
                  className='mt-2 !rounded-lg'
                />
                <div className='mt-2 flex flex-wrap gap-2'>
                  {userState?.user?.display_name ? (
                    <Tag color='green'>{t('已设置')}</Tag>
                  ) : (
                    <Tag color='grey'>{t('未设置')}</Tag>
                  )}
                  {userState?.user?.require_display_name_enabled ? (
                    <Tag color='orange'>{t('系统要求填写')}</Tag>
                  ) : null}
                </div>
              </div>
            </div>

            <div className='flex justify-end mt-4'>
              <Button
                type='primary'
                onClick={saveProfile}
                loading={profileSaving}
              >
                {t('保存用户名称')}
              </Button>
            </div>
          </Card>

          {/* 账户管理和其他设置 */}
          <div className='grid grid-cols-1 xl:grid-cols-2 items-start gap-4 md:gap-6 mt-4 md:mt-6'>
            {/* 左侧：账户管理设置 */}
            <AccountManagement
              t={t}
              userState={userState}
              status={status}
              systemToken={systemToken}
              setShowEmailBindModal={setShowEmailBindModal}
              setShowWeChatBindModal={setShowWeChatBindModal}
              generateAccessToken={generateAccessToken}
              handleSystemTokenClick={handleSystemTokenClick}
              setShowChangePasswordModal={setShowChangePasswordModal}
              setShowAccountDeleteModal={setShowAccountDeleteModal}
              passkeyStatus={passkeyStatus}
              passkeySupported={passkeySupported}
              passkeyRegisterLoading={passkeyRegisterLoading}
              passkeyDeleteLoading={passkeyDeleteLoading}
              onPasskeyRegister={handleRegisterPasskey}
              onPasskeyDelete={handleRemovePasskey}
              onBackToPayAsYouGo={backToPayAsYouGo}
              backToPayAsYouGoLoading={backToPayAsYouGoLoading}
            />

            {/* 右侧：其他设置 */}
            <NotificationSettings
              t={t}
              notificationSettings={notificationSettings}
              handleNotificationSettingChange={handleNotificationSettingChange}
              saveNotificationSettings={saveNotificationSettings}
            />
          </div>

          <SupportAccessCard
            t={t}
            supportAccessState={supportAccessState}
            supportAccessLoading={supportAccessLoading}
            onOpenSupportAccess={handleOpenSupportAccess}
            onCloseSupportAccess={handleCloseSupportAccess}
          />
        </div>
      </div>

      {/* 模态框组件 */}
      <Modal
        title={t('处理客服访问请求')}
        visible={Boolean(approvalRequest)}
        onCancel={() => {
          setApprovalRequest(null);
          clearSupportAccessToken();
        }}
        footer={
          <div className='flex justify-end gap-2'>
            <Button
              onClick={() => {
                setApprovalRequest(null);
                clearSupportAccessToken();
              }}
            >
              {t('稍后处理')}
            </Button>
            <Button
              theme='light'
              type='danger'
              loading={approvalRequestLoading}
              disabled={approvalRequest?.state !== 'pending'}
              onClick={handleRejectAccessRequest}
            >
              {t('拒绝')}
            </Button>
            <Button
              type='primary'
              loading={approvalRequestLoading}
              disabled={approvalRequest?.state !== 'pending'}
              onClick={handleApproveAccessRequest}
            >
              {t('批准访问')}
            </Button>
          </div>
        }
        width={640}
      >
        {approvalRequest ? (
          <div className='flex flex-col gap-4'>
            <Banner
              type={approvalRequest.requested_read_only ? 'info' : 'warning'}
              title={
                approvalRequest.requested_read_only
                  ? t('这是一个只读会话请求')
                  : t('这是一个标准会话请求')
              }
              description={
                approvalRequest.requested_read_only
                  ? t('批准后，客服只能查看你的账户数据，不能代表你执行操作。')
                  : t('批准后，客服可以像你本人一样查看并执行操作。')
              }
            />

            <div className='rounded-2xl border border-semi-color-border p-4'>
              <div className='flex flex-col gap-2'>
                <div>
                  <Text strong>{t('请求人')}</Text>
                  <div className='mt-1'>
                    {approvalRequest.operator_username || '-'}
                  </div>
                </div>
                <div>
                  <Text strong>{t('请求时间')}</Text>
                  <div className='mt-1'>
                    {approvalRequest.requested_at
                      ? formatDateTimeString(
                          new Date(approvalRequest.requested_at),
                        )
                      : '-'}
                  </div>
                </div>
                <div>
                  <Text strong>{t('使用限制')}</Text>
                  <div className='mt-1 text-sm text-semi-color-text-2'>
                    {t(
                      '批准后仅可激活一次，管理员必须在 24 小时内开始使用；激活后的会话会在 24 小时后失效。',
                    )}
                  </div>
                </div>
                <div>
                  <Text strong>{t('当前状态')}</Text>
                  <div className='mt-1'>
                    {approvalRequest.state === 'pending'
                      ? t('待处理')
                      : approvalRequest.state === 'approved'
                        ? t('已批准')
                        : approvalRequest.state === 'rejected'
                          ? t('已拒绝')
                          : approvalRequest.state}
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : null}
      </Modal>

      <EmailBindModal
        t={t}
        showEmailBindModal={showEmailBindModal}
        setShowEmailBindModal={setShowEmailBindModal}
        inputs={inputs}
        handleInputChange={handleInputChange}
        sendVerificationCode={sendVerificationCode}
        bindEmail={bindEmail}
        disableButton={disableButton}
        loading={loading}
        countdown={countdown}
        turnstileEnabled={turnstileEnabled}
        turnstileSiteKey={turnstileSiteKey}
        setTurnstileToken={setTurnstileToken}
      />

      <WeChatBindModal
        t={t}
        showWeChatBindModal={showWeChatBindModal}
        setShowWeChatBindModal={setShowWeChatBindModal}
        inputs={inputs}
        handleInputChange={handleInputChange}
        bindWeChat={bindWeChat}
        status={status}
      />

      <AccountDeleteModal
        t={t}
        showAccountDeleteModal={showAccountDeleteModal}
        setShowAccountDeleteModal={setShowAccountDeleteModal}
        inputs={inputs}
        handleInputChange={handleInputChange}
        deleteAccount={deleteAccount}
        userState={userState}
        turnstileEnabled={turnstileEnabled}
        turnstileSiteKey={turnstileSiteKey}
        setTurnstileToken={setTurnstileToken}
      />

      <ChangePasswordModal
        t={t}
        showChangePasswordModal={showChangePasswordModal}
        setShowChangePasswordModal={setShowChangePasswordModal}
        inputs={inputs}
        handleInputChange={handleInputChange}
        changePassword={changePassword}
        turnstileEnabled={turnstileEnabled}
        turnstileSiteKey={turnstileSiteKey}
        setTurnstileToken={setTurnstileToken}
      />

      <SecureVerificationModal
        visible={isSecureVerificationVisible}
        verificationMethods={verificationMethods}
        verificationState={verificationState}
        onVerify={executeVerification}
        onCancel={cancelVerification}
        onCodeChange={setVerificationCode}
        onMethodSwitch={switchVerificationMethod}
        title={verificationState.title || t('安全验证')}
        description={verificationState.description}
      />
    </div>
  );
};

export default PersonalSetting;
