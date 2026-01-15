import React, { useEffect, useState } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Card, Button, Space, Typography, Spin } from '@douyinfe/semi-ui';
import { API } from '../helpers';
import { showError, showSuccess } from '../helpers';

const { Title, Text } = Typography;

const SSOAuthorize = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [approving, setApproving] = useState(false);

  const nonce = searchParams.get('nonce');
  const metadata = searchParams.get('metadata');
  const postauth = searchParams.get('postauth');

  useEffect(() => {
    // 验证必需参数
    if (!nonce || !postauth) {
      showError('缺少必需参数');
      navigate('/');
    }
  }, [nonce, postauth, navigate]);

  const handleApprove = async () => {
    setApproving(true);
    try {
      const response = await API.post('/api/sso-beta/approve', {
        nonce,
        metadata,
        postauth
      });

      if (response.data.success) {
        const redirectUrl = response.data.data.redirect_url;
        window.location.href = redirectUrl;
      } else {
        showError(response.data.message || '授权失败');
        setApproving(false);
      }
    } catch (error) {
      showError('授权失败：' + error.message);
      setApproving(false);
    }
  };

  const handleCancel = async () => {
    try {
      const response = await API.post('/api/sso-beta/cancel');
      if (response.data.success) {
        const redirectUrl = response.data.data.redirect_url;
        navigate(redirectUrl);
      }
    } catch (error) {
      navigate('/');
    }
  };

  if (loading) {
    return (
      <div style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh'
      }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      background: 'var(--semi-color-bg-0)',
      padding: '20px'
    }}>
      <Card
        style={{
          maxWidth: '500px',
          width: '100%'
        }}
        bodyStyle={{
          padding: '32px'
        }}
      >
        <Space vertical align="start" spacing="medium" style={{ width: '100%' }}>
          <Title heading={3} style={{ margin: 0 }}>
            授权确认
          </Title>

          <Text style={{ fontSize: '16px', marginTop: '8px' }}>
            是否授权 <strong>Privnode 支持</strong> 访问您的账号信息？
          </Text>

          <Card
            style={{
              width: '100%',
              background: 'var(--semi-color-fill-0)',
              marginTop: '16px'
            }}
          >
            <Space vertical spacing="small">
              <Text strong>Privnode 支持 将可以：</Text>
              <Text>• 获取您的用户基本信息</Text>
            </Space>
          </Card>

          <Text type="secondary" style={{ fontSize: '14px', marginTop: '8px' }}>
            他们无法代表您执行操作。
          </Text>

          <Space style={{ width: '100%', justifyContent: 'center', marginTop: '24px' }}>
            <Button
              theme="solid"
              type="primary"
              size="large"
              loading={approving}
              onClick={handleApprove}
              style={{ minWidth: '120px' }}
            >
              授权
            </Button>
            <Button
              size="large"
              onClick={handleCancel}
              disabled={approving}
              style={{ minWidth: '120px' }}
            >
              取消
            </Button>
          </Space>
        </Space>
      </Card>
    </div>
  );
};

export default SSOAuthorize;
