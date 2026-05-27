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

import React, { useEffect, useMemo, useState } from 'react';
import {
  Banner,
  Button,
  Card,
  Form,
  Input,
  Typography,
} from '@douyinfe/semi-ui';
import { CheckCircle2, LockKeyhole, LogIn } from 'lucide-react';
import { useLocation, useParams } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';

const { Text, Title } = Typography;

function parseMigrationHash(hash) {
  const clean = (hash || '').replace(/^#/, '').replace(/^~\//, '');
  const parts = clean.split('/').filter(Boolean);
  if (parts[0] === 't') {
    return {
      mode: 'target',
      migrationToken: parts[1] || '',
      accessToken: parts[2] || '',
      userToken: parts[5] || '',
    };
  }
  if (parts[0] === 'i') {
    return {
      mode: 'import',
      setupToken: parts[1] || '',
      accessToken: parts[2] || '',
    };
  }
  return { mode: 'unknown' };
}

const Migrate = () => {
  const location = useLocation();
  const params = useParams();
  const parsed = useMemo(
    () => parseMigrationHash(location.hash),
    [location.hash],
  );
  const [loading, setLoading] = useState(false);
  const [verifyResult, setVerifyResult] = useState(null);
  const [loginValue, setLoginValue] = useState({ username: '', password: '' });
  const [passwordValue, setPasswordValue] = useState({
    password: '',
    password2: '',
  });
  const [complete, setComplete] = useState(false);

  const migrateID = params.migrateId;

  const verifyTarget = async () => {
    if (parsed.mode !== 'target' || !migrateID) return;
    setLoading(true);
    try {
      const res = await API.post(
        `/migrate/api/migrations/${migrateID}/verify`,
        {
          migration_token: parsed.migrationToken,
          access_token: parsed.accessToken,
          user_token: parsed.userToken,
        },
      );
      if (res.data.success) {
        setVerifyResult(res.data.data);
        setComplete(Boolean(res.data.data?.captured));
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const verifyImport = async () => {
    if (parsed.mode !== 'import') return;
    setLoading(true);
    try {
      const res = await API.post('/migrate/api/imports/setup/verify', {
        setup_token: parsed.setupToken,
        access_token: parsed.accessToken,
      });
      if (res.data.success) {
        setVerifyResult(res.data.data);
        setComplete(res.data.data?.status === 'active');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (parsed.mode === 'target') {
      verifyTarget();
    } else if (parsed.mode === 'import') {
      verifyImport();
    }
  }, [
    parsed.mode,
    migrateID,
    parsed.migrationToken,
    parsed.setupToken,
    parsed.accessToken,
    parsed.userToken,
  ]);

  const captureCurrent = async () => {
    setLoading(true);
    try {
      const res = await API.post(
        `/migrate/api/migrations/${migrateID}/capture`,
        {
          migration_token: parsed.migrationToken,
          access_token: parsed.accessToken,
          user_token: parsed.userToken,
        },
      );
      if (res.data.success) {
        setVerifyResult(res.data.data);
        setComplete(true);
        showSuccess('迁移数据已记录');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const loginAndCapture = async () => {
    setLoading(true);
    try {
      const res = await API.post(`/migrate/api/migrations/${migrateID}/login`, {
        migration_token: parsed.migrationToken,
        access_token: parsed.accessToken,
        user_token: parsed.userToken,
        username: loginValue.username,
        password: loginValue.password,
      });
      if (res.data.success) {
        setComplete(true);
        setVerifyResult(res.data.data?.migration || res.data.data);
        showSuccess('迁移数据已记录');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const setPassword = async () => {
    if (passwordValue.password !== passwordValue.password2) {
      showError('两次输入的密码不一致');
      return;
    }
    setLoading(true);
    try {
      const res = await API.post('/migrate/api/imports/setup/password', {
        setup_token: parsed.setupToken,
        access_token: parsed.accessToken,
        password: passwordValue.password,
      });
      if (res.data.success) {
        setComplete(true);
        setVerifyResult(res.data.data);
        showSuccess('密码已设置');
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const renderTargetFlow = () => {
    if (complete) {
      return (
        <Banner
          type='success'
          title='迁移确认已完成'
          description='系统已记录迁移数据。接收方确认导入成功后，原账户会以“已迁移”为原因禁用。'
        />
      );
    }

    return (
      <div className='space-y-4'>
        {verifyResult && (
          <Banner
            type='info'
            title={`请登录指定账户：${verifyResult.cah_id}`}
            description={`迁移数据会绑定到邮箱 ${verifyResult.email}`}
          />
        )}
        {verifyResult?.login_ok ? (
          <Button
            type='primary'
            icon={<CheckCircle2 size={16} />}
            loading={loading}
            onClick={captureCurrent}
          >
            使用当前登录账户确认迁移
          </Button>
        ) : (
          <Form labelPosition='top'>
            <Form.Input
              label='用户名'
              value={loginValue.username}
              onChange={(value) =>
                setLoginValue((prev) => ({ ...prev, username: value }))
              }
            />
            <Form.Input
              label='密码'
              mode='password'
              value={loginValue.password}
              onChange={(value) =>
                setLoginValue((prev) => ({ ...prev, password: value }))
              }
            />
            <Button
              type='primary'
              icon={<LogIn size={16} />}
              loading={loading}
              onClick={loginAndCapture}
            >
              登录并确认迁移
            </Button>
          </Form>
        )}
      </div>
    );
  };

  const renderImportFlow = () => {
    if (complete) {
      return (
        <Banner
          type='success'
          title='账户已启用'
          description='密码已经设置完成，可以使用迁移后的账户登录。'
        />
      );
    }

    return (
      <div className='space-y-4'>
        {verifyResult && (
          <Banner
            type='info'
            title={`设置迁移账户密码：${verifyResult.cah_id}`}
            description={`账户邮箱：${verifyResult.email}`}
          />
        )}
        <Form labelPosition='top'>
          <Form.Input
            label='新密码'
            mode='password'
            value={passwordValue.password}
            onChange={(value) =>
              setPasswordValue((prev) => ({ ...prev, password: value }))
            }
          />
          <Form.Input
            label='确认新密码'
            mode='password'
            value={passwordValue.password2}
            onChange={(value) =>
              setPasswordValue((prev) => ({ ...prev, password2: value }))
            }
          />
          <Button
            type='primary'
            icon={<LockKeyhole size={16} />}
            loading={loading}
            onClick={setPassword}
          >
            设置密码并启用账户
          </Button>
        </Form>
      </div>
    );
  };

  return (
    <main className='min-h-screen bg-[var(--semi-color-bg-0)] px-4 py-10'>
      <div className='mx-auto max-w-[560px]'>
        <div className='mb-6'>
          <Title heading={3}>PrivHub 账户迁移</Title>
          <Text type='secondary'>
            此页面仅处理 `/migrate` 路径下的迁移流程。
          </Text>
        </div>
        <Card>
          {parsed.mode === 'target' && renderTargetFlow()}
          {parsed.mode === 'import' && renderImportFlow()}
          {parsed.mode === 'unknown' && (
            <Banner
              type='danger'
              title='迁移链接无效'
              description='链接缺少必要的迁移参数。'
            />
          )}
        </Card>
      </div>
    </main>
  );
};

export default Migrate;
