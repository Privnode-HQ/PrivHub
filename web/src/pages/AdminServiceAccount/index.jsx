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

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Input,
  InputNumber,
  Modal,
  Space,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import {
  Copy,
  KeyRound,
  Plus,
  RotateCcw,
  ShieldCheck,
  Trash2,
} from 'lucide-react';
import CardPro from '../../components/common/ui/CardPro';
import CardTable from '../../components/common/ui/CardTable';
import {
  API,
  copy,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';

const { Text } = Typography;

const emptyForm = {
  name: '',
  description: '',
  target: '',
  expires_in_days: 90,
  allow_ips: '',
};

const formatTime = (value) => {
  if (!value || value <= 0) {
    return '-';
  }
  return timestamp2string(value);
};

const renderStatus = (record, t) => {
  if (record.expired) {
    return <Tag color='orange'>{t('已过期')}</Tag>;
  }
  if (record.status === 1) {
    return <Tag color='green'>{t('已启用')}</Tag>;
  }
  if (record.status === 2) {
    return <Tag color='red'>{t('已停用')}</Tag>;
  }
  return <Tag color='grey'>{t('未知状态')}</Tag>;
};

const AdminServiceAccount = () => {
  const { t } = useTranslation();
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [total, setTotal] = useState(0);
  const [createVisible, setCreateVisible] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [saving, setSaving] = useState(false);
  const [credentialData, setCredentialData] = useState(null);

  const loadAccounts = useCallback(
    async (page = activePage, size = pageSize, search = keyword) => {
      setLoading(true);
      try {
        const res = await API.get('/api/admin/service-accounts/', {
          params: {
            p: page,
            page_size: size,
            keyword: search.trim(),
          },
        });
        const { success, message, data } = res.data;
        if (!success) {
          showError(message);
          return;
        }
        setAccounts(data.items || []);
        setActivePage(data.page <= 0 ? 1 : data.page);
        setTotal(data.total || 0);
      } catch (error) {
        showError(error.message);
      } finally {
        setLoading(false);
      }
    },
    [activePage, keyword, pageSize],
  );

  useEffect(() => {
    loadAccounts(1, pageSize, '');
  }, []);

  const updateForm = (key, value) => {
    setForm((current) => ({
      ...current,
      [key]: value,
    }));
  };

  const handleSearch = () => {
    setActivePage(1);
    loadAccounts(1, pageSize, keyword);
  };

  const handleCreate = async () => {
    if (!form.name.trim()) {
      showError(t('请输入 Service Account 名称'));
      return;
    }
    setSaving(true);
    try {
      const res = await API.post('/api/admin/service-accounts/', {
        ...form,
        expires_in_days: Number(form.expires_in_days) || 90,
        allow_ips: form.allow_ips,
      });
      const { success, message, data } = res.data;
      if (!success) {
        showError(message);
        return;
      }
      showSuccess(t('创建成功'));
      setCreateVisible(false);
      setForm(emptyForm);
      setCredentialData(data);
      await loadAccounts(1, pageSize, keyword);
    } catch (error) {
      showError(error.message);
    } finally {
      setSaving(false);
    }
  };

  const updateAccountStatus = async (record, status) => {
    setLoading(true);
    try {
      const res = await API.put(`/api/admin/service-accounts/${record.id}`, {
        name: record.name,
        description: record.description || '',
        status,
        allow_ips: record.allow_ips || '',
      });
      const { success, message } = res.data;
      if (!success) {
        showError(message);
        return;
      }
      showSuccess(t('操作成功'));
      await loadAccounts(activePage, pageSize, keyword);
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const rotateCredential = (record) => {
    Modal.confirm({
      title: t('轮换凭据'),
      content: t('旧 JWT 会立即失效，系统将生成新的 Bearer JWT。'),
      okText: t('轮换'),
      cancelText: t('取消'),
      onOk: async () => {
        setLoading(true);
        try {
          const res = await API.post(
            `/api/admin/service-accounts/${record.id}/rotate`,
            { expires_in_days: 90 },
          );
          const { success, message, data } = res.data;
          if (!success) {
            showError(message);
            return;
          }
          showSuccess(t('凭据已轮换'));
          setCredentialData(data);
          await loadAccounts(activePage, pageSize, keyword);
        } catch (error) {
          showError(error.message);
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const deleteAccount = (record) => {
    Modal.confirm({
      title: t('删除 Service Account'),
      content: t('删除后该 JWT 凭据会立即失效。'),
      okText: t('删除'),
      cancelText: t('取消'),
      okButtonProps: { type: 'danger' },
      onOk: async () => {
        setLoading(true);
        try {
          const res = await API.delete(
            `/api/admin/service-accounts/${record.id}`,
          );
          const { success, message } = res.data;
          if (!success) {
            showError(message);
            return;
          }
          showSuccess(t('删除成功'));
          await loadAccounts(activePage, pageSize, keyword);
        } catch (error) {
          showError(error.message);
        } finally {
          setLoading(false);
        }
      },
    });
  };

  const columns = useMemo(
    () => [
      {
        title: t('Service Account'),
        dataIndex: 'name',
        render: (_, record) => (
          <div className='flex flex-col gap-1'>
            <Text strong>{record.name}</Text>
            <Text type='secondary' size='small' copyable>
              {record.service_account_id}
            </Text>
            {record.description ? (
              <Text type='secondary' size='small'>
                {record.description}
              </Text>
            ) : null}
          </div>
        ),
      },
      {
        title: t('绑定管理员'),
        dataIndex: 'target',
        render: (_, record) => (
          <div className='flex flex-col gap-1'>
            <Text>{record.target?.username || record.username || '-'}</Text>
            <Text type='secondary' size='small'>
              {record.target?.cah_id || record.user_cah_id || '-'}
            </Text>
          </div>
        ),
      },
      {
        title: t('状态'),
        dataIndex: 'status',
        render: (_, record) => renderStatus(record, t),
      },
      {
        title: t('IP 白名单'),
        dataIndex: 'allow_ips',
        render: (value) => {
          const count = value ? value.split(/\s+/).filter(Boolean).length : 0;
          return count > 0 ? (
            <Tag color='blue'>{t('{{count}} 条', { count })}</Tag>
          ) : (
            <Text type='secondary'>{t('未限制')}</Text>
          );
        },
      },
      {
        title: t('最后使用'),
        dataIndex: 'accessed_time',
        render: (value) => formatTime(value),
      },
      {
        title: t('过期时间'),
        dataIndex: 'expires_at',
        render: (value) => formatTime(value),
      },
      {
        title: t('操作'),
        dataIndex: 'operate',
        render: (_, record) => (
          <Space wrap>
            <Button
              size='small'
              type='tertiary'
              icon={<RotateCcw size={14} />}
              onClick={() => rotateCredential(record)}
            >
              {t('轮换')}
            </Button>
            {record.status === 1 ? (
              <Button
                size='small'
                type='warning'
                onClick={() =>
                  Modal.confirm({
                    title: t('停用 Service Account'),
                    content: t('停用后该 JWT 会立即失效。'),
                    okText: t('停用'),
                    cancelText: t('取消'),
                    onOk: () => updateAccountStatus(record, 2),
                  })
                }
              >
                {t('停用')}
              </Button>
            ) : (
              <Button
                size='small'
                type='tertiary'
                onClick={() => updateAccountStatus(record, 1)}
              >
                {t('启用')}
              </Button>
            )}
            <Button
              size='small'
              type='danger'
              icon={<Trash2 size={14} />}
              onClick={() => deleteAccount(record)}
            >
              {t('删除')}
            </Button>
          </Space>
        ),
      },
    ],
    [activePage, keyword, pageSize, t],
  );

  return (
    <div className='mt-[60px] px-2'>
      <CardPro
        type='type1'
        descriptionArea={
          <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
            <div className='flex items-center text-blue-500'>
              <ShieldCheck size={18} className='mr-2' />
              <Text>{t('Admin Service Account')}</Text>
            </div>
            <Text type='secondary' className='text-sm'>
              {t('为管理员生成可撤销、可轮换、带过期时间的 Bearer JWT')}
            </Text>
          </div>
        }
        actionsArea={
          <div className='flex flex-col md:flex-row gap-2 w-full'>
            <Input
              value={keyword}
              onChange={setKeyword}
              placeholder={t('搜索名称、ASA ID、用户名或 CAH ID')}
              prefix={<IconSearch />}
              showClear
              onKeyDown={(event) => {
                if (event.key === 'Enter') {
                  handleSearch();
                }
              }}
            />
            <Space>
              <Button loading={loading} onClick={handleSearch}>
                {t('查询')}
              </Button>
              <Button
                theme='solid'
                type='primary'
                icon={<Plus size={14} />}
                onClick={() => setCreateVisible(true)}
              >
                {t('新建')}
              </Button>
            </Space>
          </div>
        }
        t={t}
      >
        <CardTable
          columns={columns}
          dataSource={accounts}
          rowKey='id'
          loading={loading}
          scroll={{ x: 'max-content' }}
          pagination={{
            currentPage: activePage,
            pageSize,
            total,
            showSizeChanger: true,
            pageSizeOptions: [10, 20, 50, 100],
            onPageSizeChange: (size) => {
              setPageSize(size);
              setActivePage(1);
              loadAccounts(1, size, keyword);
            },
            onPageChange: (page) => {
              setActivePage(page);
              loadAccounts(page, pageSize, keyword);
            },
          }}
          className='rounded-lg overflow-hidden'
        />
      </CardPro>

      <Modal
        title={t('新建 Admin Service Account')}
        visible={createVisible}
        onOk={handleCreate}
        onCancel={() => setCreateVisible(false)}
        okText={t('生成 JWT')}
        cancelText={t('取消')}
        confirmLoading={saving}
        width={640}
      >
        <div className='flex flex-col gap-3'>
          <Input
            value={form.name}
            onChange={(value) => updateForm('name', value)}
            placeholder={t('名称，例如 CI 部署机器人')}
            prefix={<KeyRound size={14} />}
          />
          <Input
            value={form.target}
            onChange={(value) => updateForm('target', value)}
            placeholder={t(
              '目标管理员：留空为自己，或输入用户 ID / CAH ID / 用户名',
            )}
          />
          <Input
            value={form.description}
            onChange={(value) => updateForm('description', value)}
            placeholder={t('说明，可选')}
          />
          <div className='grid grid-cols-1 md:grid-cols-2 gap-3'>
            <InputNumber
              value={form.expires_in_days}
              onChange={(value) => updateForm('expires_in_days', value)}
              min={1}
              max={365}
              suffix={t('天')}
            />
            <Text type='secondary' className='text-sm self-center'>
              {t('最长 365 天，到期后 JWT 自动失效')}
            </Text>
          </div>
          <TextArea
            value={form.allow_ips}
            onChange={(value) => updateForm('allow_ips', value)}
            placeholder={t('IP 白名单，可选。每行一个 IP 或 CIDR')}
            autosize={{ minRows: 3, maxRows: 6 }}
          />
        </div>
      </Modal>

      <Modal
        title={t('JWT 凭据')}
        visible={!!credentialData}
        onCancel={() => setCredentialData(null)}
        footer={
          <Space>
            <Button onClick={() => setCredentialData(null)}>{t('关闭')}</Button>
            <Button
              theme='solid'
              type='primary'
              icon={<Copy size={14} />}
              onClick={async () => {
                const ok = await copy(credentialData?.credential || '');
                if (ok) {
                  showSuccess(t('已复制到剪贴板'));
                }
              }}
            >
              {t('复制')}
            </Button>
          </Space>
        }
        width={720}
      >
        <div className='flex flex-col gap-3'>
          <Tag color='orange'>{t('该 JWT 仅显示一次，请立即保存')}</Tag>
          <TextArea
            value={credentialData?.credential || ''}
            readOnly
            autosize={{ minRows: 8, maxRows: 12 }}
          />
          <Text type='secondary' className='text-sm'>
            {t(
              '调用 API 时使用 Authorization: Bearer <JWT>，不需要 New-Api-User 请求头',
            )}
          </Text>
        </div>
      </Modal>
    </div>
  );
};

export default AdminServiceAccount;
