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

import React, { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Empty,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IllustrationNoContent,
  IllustrationNoContentDark,
} from '@douyinfe/semi-illustrations';
import {
  Copy,
  Eye,
  FilePlus2,
  RefreshCw,
  Rocket,
  Save,
  Search,
  Trash2,
} from 'lucide-react';
import {
  API,
  formatDateTimeString,
  getRelativeTime,
  isSupport,
  showError,
  showSuccess,
} from '../../helpers';
import { useTranslation } from 'react-i18next';

const PAGE_SIZE = 10;
const { Text } = Typography;

const emptyEditor = {
  title: '',
  content: '',
  target_type: 'all',
  target_groups: [],
  target_user_ids: [],
};

const targetTypeOptions = [
  { label: '全体用户', value: 'all' },
  { label: '指定分组', value: 'groups' },
  { label: '指定用户', value: 'users' },
];

const AdminMessages = () => {
  const { t } = useTranslation();
  const readOnlyAdmin = isSupport();
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [savingTemplate, setSavingTemplate] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [template, setTemplate] = useState('');
  const [placeholders, setPlaceholders] = useState([]);
  const [editorVisible, setEditorVisible] = useState(false);
  const [previewMessage, setPreviewMessage] = useState(null);
  const [editingMessage, setEditingMessage] = useState(null);
  const [editorValue, setEditorValue] = useState(emptyEditor);
  const [submitting, setSubmitting] = useState(false);
  const [groupOptions, setGroupOptions] = useState([]);
  const [userOptions, setUserOptions] = useState([]);
  const [userSearchKeyword, setUserSearchKeyword] = useState('');
  const [userSearchLoading, setUserSearchLoading] = useState(false);

  const loadMessages = async (targetPage = page) => {
    setLoading(true);
    try {
      const res = await API.get('/api/message', {
        params: {
          p: targetPage,
          page_size: PAGE_SIZE,
        },
      });
      if (res.data.success) {
        setMessages(res.data.data.items || []);
        setTotal(res.data.data.total || 0);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const loadTemplate = async () => {
    try {
      const res = await API.get('/api/message/template');
      if (res.data.success) {
        setTemplate(res.data.data.template || '');
        setPlaceholders(res.data.data.placeholders || []);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const loadGroups = async () => {
    try {
      const res = await API.get('/api/group/');
      if (res.data.success) {
        const options = (res.data.data || [])
          .map((group) => ({
            label: group,
            value: group,
          }))
          .sort((a, b) => a.label.localeCompare(b.label));
        setGroupOptions(options);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const mergeUserOptions = (incomingUsers = []) => {
    setUserOptions((prev) => {
      const merged = [...prev];
      const existingIds = new Set(prev.map((item) => item.value));

      incomingUsers.forEach((user) => {
        if (!user?.id || existingIds.has(user.id)) {
          return;
        }
        const userLabel = user.display_name || user.username;
        merged.push({
          label: `${userLabel}${user.cah_id ? ` (${user.cah_id})` : ` (#${user.id})`}${user.group ? ` / ${user.group}` : ''}${user.email ? ` / ${user.email}` : ''}`,
          value: user.id,
        });
        existingIds.add(user.id);
      });

      return merged;
    });
  };

  const searchUsers = async (keyword = userSearchKeyword) => {
    setUserSearchLoading(true);
    try {
      const res = await API.get('/api/user/search', {
        params: {
          keyword,
          p: 1,
          page_size: 20,
        },
      });

      if (res.data.success) {
        mergeUserOptions(res.data.data.items || []);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setUserSearchLoading(false);
    }
  };

  useEffect(() => {
    loadMessages(page);
  }, [page]);

  useEffect(() => {
    loadTemplate();
    loadGroups();
  }, []);

  const openCreateModal = () => {
    setEditingMessage(null);
    setEditorValue({ ...emptyEditor });
    setUserSearchKeyword('');
    setEditorVisible(true);
  };

  const openEditModal = (record) => {
    setEditingMessage(record);
    setEditorValue({
      title: record.title || '',
      content: record.content || '',
      target_type: record.target_type || 'all',
      target_groups: record.target_groups || [],
      target_user_ids: record.target_user_ids || [],
    });
    mergeUserOptions(record.target_user_options || []);
    setUserSearchKeyword('');
    setEditorVisible(true);
  };

  const duplicateMessage = async (record) => {
    setSubmitting(true);
    try {
      const res = await API.post(`/api/message/${record.id}/copy`);
      if (res.data.success) {
        showSuccess(t('草稿已复制'));
        setPage(1);
        await loadMessages(1);
        openEditModal({
          ...(res.data.data || {}),
          target_type:
            res.data.data?.target_type || record.target_type || 'all',
          target_groups:
            res.data.data?.target_groups || record.target_groups || [],
          target_user_ids:
            res.data.data?.target_user_ids || record.target_user_ids || [],
          target_user_options: record.target_user_options || [],
          status: 'draft',
        });
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const submitDraft = async () => {
    if (!editorValue.title.trim() || !editorValue.content.trim()) {
      showError(t('标题和内容不能为空'));
      return;
    }

    setSubmitting(true);
    try {
      const res = editingMessage
        ? await API.put(`/api/message/${editingMessage.id}`, editorValue)
        : await API.post('/api/message', editorValue);

      if (res.data.success) {
        showSuccess(editingMessage ? t('草稿已更新') : t('草稿已创建'));
        setEditorVisible(false);
        if (editingMessage) {
          await loadMessages(page);
        } else {
          setPage(1);
          await loadMessages(1);
        }
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const publishMessage = async (record) => {
    setSubmitting(true);
    try {
      const res = await API.post(`/api/message/${record.id}/publish`);
      if (res.data.success) {
        showSuccess(t('消息已上线'));
        await loadMessages(page);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const retryDelivery = async (record) => {
    setSubmitting(true);
    try {
      const res = await API.post(`/api/message/${record.id}/retry`);
      if (res.data.success) {
        showSuccess(
          `${t('重试任务已提交')} (${res.data.data || 0} ${t('封邮件')})`,
        );
        await loadMessages(page);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const deleteMessage = async (record) => {
    setSubmitting(true);
    try {
      const res = await API.delete(`/api/message/${record.id}`);
      if (res.data.success) {
        showSuccess(t('草稿已删除'));
        await loadMessages(page);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSubmitting(false);
    }
  };

  const saveTemplate = async () => {
    setSavingTemplate(true);
    try {
      const res = await API.put('/api/message/template', {
        template,
      });
      if (res.data.success) {
        showSuccess(t('模板已更新'));
        await loadMessages(page);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSavingTemplate(false);
    }
  };

  const columns = [
    {
      title: t('标题'),
      dataIndex: 'title',
      render: (text) => <span className='font-medium'>{text}</span>,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 120,
      render: (value) => (
        <Tag color={value === 'online' ? 'green' : 'grey'}>
          {value === 'online' ? t('已上线') : t('草稿')}
        </Tag>
      ),
    },
    {
      title: t('发送范围'),
      width: 220,
      render: (_, record) => {
        if (record.target_type === 'groups') {
          return (
            <div>
              <Tag color='blue'>{t('指定分组')}</Tag>
              <div className='mt-1 text-sm text-semi-color-text-1'>
                {(record.target_groups || []).join(', ') || '-'}
              </div>
            </div>
          );
        }

        if (record.target_type === 'users') {
          return (
            <div>
              <Tag color='orange'>{t('指定用户')}</Tag>
              <div className='mt-1 text-sm text-semi-color-text-1'>
                {(record.target_user_options || [])
                  .map(
                    (user) =>
                      `${user.display_name || user.username}${user.cah_id ? ` (${user.cah_id})` : ''}`,
                  )
                  .join(', ') || '-'}
              </div>
            </div>
          );
        }

        return <Tag color='green'>{t('全体用户')}</Tag>;
      },
    },
    {
      title: t('时间'),
      dataIndex: 'published_at',
      width: 220,
      render: (value, record) => {
        const dateValue = value || record.created_at;
        const date = dateValue ? new Date(dateValue) : null;
        return (
          <div>
            <div>{date ? formatDateTimeString(date) : '-'}</div>
            <Text type='secondary'>
              {dateValue ? getRelativeTime(dateValue) : '-'}
            </Text>
          </div>
        );
      },
    },
    {
      title: t('投递'),
      width: 180,
      render: (_, record) => (
        <div>
          <div>
            {t('站内')}: {record.delivery_total || 0}
          </div>
          <Text type='secondary'>
            {t('已读')}: {record.read_total || 0} / {record.delivery_total || 0}
          </Text>
          <div>
            <Text type='secondary'>
              {t('邮件')}: {record.email_sent || 0}
            </Text>
          </div>
          <div>
            <Text type='secondary'>
              {t('失败')}: {record.email_failed || 0}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: t('操作'),
      width: 280,
      render: (_, record) => (
        <Space wrap>
          <Button
            icon={<Eye size={14} />}
            theme='light'
            onClick={() => setPreviewMessage(record)}
          >
            {t('预览')}
          </Button>
          {!readOnlyAdmin && (
            <Button
              icon={<Copy size={14} />}
              theme='light'
              loading={submitting}
              onClick={() => duplicateMessage(record)}
            >
              {t('复制')}
            </Button>
          )}

          {!readOnlyAdmin && record.status === 'online' && (
            <Button
              icon={<RefreshCw size={14} />}
              theme='light'
              loading={submitting}
              disabled={!record.email_failed}
              onClick={() => retryDelivery(record)}
            >
              {t('重试投递')}
            </Button>
          )}

          {!readOnlyAdmin && record.status === 'draft' && (
            <>
              <Button
                icon={<Save size={14} />}
                theme='light'
                onClick={() => openEditModal(record)}
              >
                {t('编辑')}
              </Button>
              <Button
                icon={<Rocket size={14} />}
                type='primary'
                loading={submitting}
                onClick={() => publishMessage(record)}
              >
                {t('上线')}
              </Button>
              <Button
                icon={<Trash2 size={14} />}
                theme='light'
                type='danger'
                loading={submitting}
                onClick={() => deleteMessage(record)}
              >
                {t('删除')}
              </Button>
            </>
          )}
        </Space>
      ),
    },
  ];

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto w-full max-w-7xl space-y-4'>
        <Card
          title={t('信息管理')}
          headerExtraContent={
            <Space>
              <Button theme='light' onClick={() => loadMessages(page)}>
                {t('刷新')}
              </Button>
              {!readOnlyAdmin && (
                <Button
                  icon={<FilePlus2 size={14} />}
                  type='primary'
                  onClick={openCreateModal}
                >
                  {t('新增消息')}
                </Button>
              )}
            </Space>
          }
        >
          <Table
            columns={columns}
            dataSource={messages}
            rowKey='id'
            loading={loading}
            pagination={{
              currentPage: page,
              pageSize: PAGE_SIZE,
              total,
              onPageChange: setPage,
            }}
            empty={
              <Empty
                image={
                  <IllustrationNoContent style={{ width: 150, height: 150 }} />
                }
                darkModeImage={
                  <IllustrationNoContentDark
                    style={{ width: 150, height: 150 }}
                  />
                }
                description={t('暂无消息')}
                style={{ padding: 24 }}
              />
            }
          />
        </Card>

        <Card
          title={t('模板设置')}
          headerExtraContent={
            !readOnlyAdmin ? (
              <Button
                type='primary'
                loading={savingTemplate}
                onClick={saveTemplate}
              >
                {t('保存模板')}
              </Button>
            ) : null
          }
        >
          <Space
            vertical
            align='start'
            spacing='tight'
            style={{ width: '100%' }}
          >
            <Text type='secondary'>
              {t(
                '站内详情和邮件都会使用这个 HTML 模板。模板中必须保留 {{.ContentHTML}} 占位符。',
              )}
            </Text>
            {placeholders.length > 0 && (
              <Text type='secondary'>
                {t('可用变量')}: {placeholders.join(', ')}
              </Text>
            )}
          </Space>

          <TextArea
            value={template}
            autosize={{ minRows: 14, maxRows: 24 }}
            style={{ marginTop: 16, fontFamily: 'JetBrains Mono, Consolas' }}
            readOnly={readOnlyAdmin}
            onChange={setTemplate}
          />
        </Card>
      </div>

      <Modal
        title={editingMessage ? t('编辑草稿') : t('新增消息')}
        visible={!readOnlyAdmin && editorVisible}
        onCancel={() => setEditorVisible(false)}
        onOk={submitDraft}
        okText={editingMessage ? t('保存草稿') : t('创建草稿')}
        confirmLoading={submitting}
        size='large'
      >
        <Space vertical align='start' style={{ width: '100%' }}>
          <Input
            value={editorValue.title}
            placeholder={t('请输入标题')}
            onChange={(value) =>
              setEditorValue((prev) => ({ ...prev, title: value }))
            }
          />
          <Select
            value={editorValue.target_type}
            optionList={targetTypeOptions.map((option) => ({
              label: t(option.label),
              value: option.value,
            }))}
            style={{ width: '100%' }}
            onChange={(value) =>
              setEditorValue((prev) => ({
                ...prev,
                target_type: value,
                target_groups: value === 'groups' ? prev.target_groups : [],
                target_user_ids: value === 'users' ? prev.target_user_ids : [],
              }))
            }
          />

          {editorValue.target_type === 'groups' && (
            <Select
              multiple
              value={editorValue.target_groups}
              optionList={groupOptions}
              placeholder={t('请选择一个或多个分组')}
              style={{ width: '100%' }}
              onChange={(value) =>
                setEditorValue((prev) => ({
                  ...prev,
                  target_groups: value || [],
                }))
              }
            />
          )}

          {editorValue.target_type === 'users' && (
            <Space vertical align='start' style={{ width: '100%' }}>
              <Space style={{ width: '100%' }}>
                <Input
                  value={userSearchKeyword}
                  prefix={<Search size={14} />}
                  placeholder={t('输入用户名、邮箱或 ID 搜索用户')}
                  onChange={setUserSearchKeyword}
                  onEnterPress={() => searchUsers(userSearchKeyword)}
                />
                <Button
                  loading={userSearchLoading}
                  onClick={() => searchUsers(userSearchKeyword)}
                >
                  {t('搜索')}
                </Button>
              </Space>
              <Select
                multiple
                value={editorValue.target_user_ids}
                optionList={userOptions}
                placeholder={t('请选择一个或多个用户')}
                style={{ width: '100%' }}
                onChange={(value) =>
                  setEditorValue((prev) => ({
                    ...prev,
                    target_user_ids: value || [],
                  }))
                }
              />
            </Space>
          )}
          <TextArea
            value={editorValue.content}
            autosize={{ minRows: 12, maxRows: 20 }}
            placeholder={t('请输入内容（支持 Markdown）')}
            onChange={(value) =>
              setEditorValue((prev) => ({ ...prev, content: value }))
            }
          />
        </Space>
      </Modal>

      <Modal
        title={previewMessage?.title || t('消息预览')}
        visible={Boolean(previewMessage)}
        footer={null}
        onCancel={() => setPreviewMessage(null)}
        size='large'
      >
        {previewMessage && (
          <div className='space-y-3'>
            <div>
              {previewMessage.target_type === 'groups' && (
                <Text type='secondary'>
                  {t('发送到分组')}:{' '}
                  {(previewMessage.target_groups || []).join(', ')}
                </Text>
              )}
              {previewMessage.target_type === 'users' && (
                <Text type='secondary'>
                  {t('发送到用户')}:{' '}
                  {(previewMessage.target_user_options || [])
                    .map(
                      (user) =>
                        `${user.display_name || user.username}${user.cah_id ? ` (${user.cah_id})` : ''}`,
                    )
                    .join(', ')}
                </Text>
              )}
              {(previewMessage.target_type || 'all') === 'all' && (
                <Text type='secondary'>{t('发送到全体用户')}</Text>
              )}
            </div>
            <div
              className='overflow-auto rounded-xl border border-semi-color-border bg-semi-color-bg-0'
              dangerouslySetInnerHTML={{
                __html: previewMessage.html_content || '',
              }}
            />
          </div>
        )}
      </Modal>
    </div>
  );
};

export default AdminMessages;
