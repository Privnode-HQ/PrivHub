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
  Modal,
  Space,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IllustrationNoContent,
  IllustrationNoContentDark,
} from '@douyinfe/semi-illustrations';
import { MailOpen, MailPlus, Eye } from 'lucide-react';
import {
  API,
  formatDateTimeString,
  getRelativeTime,
  showError,
} from '../../helpers';
import {
  MESSAGE_UNREAD_REFRESH_EVENT,
  refreshUnreadMessages,
} from '../../hooks/common/useNotifications';
import { useTranslation } from 'react-i18next';

const PAGE_SIZE = 10;
const { Text } = Typography;

const Messages = () => {
  const { t } = useTranslation();
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [selectedMessage, setSelectedMessage] = useState(null);
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);

  const loadMessages = async (targetPage = page) => {
    setLoading(true);
    try {
      const res = await API.get('/api/message/self', {
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

  useEffect(() => {
    loadMessages(page);
  }, [page]);

  useEffect(() => {
    const handleRefresh = () => {
      loadMessages(page);
    };

    window.addEventListener(MESSAGE_UNREAD_REFRESH_EVENT, handleRefresh);
    return () => {
      window.removeEventListener(MESSAGE_UNREAD_REFRESH_EVENT, handleRefresh);
    };
  }, [page]);

  const handleOpenMessage = async (message) => {
    setSelectedMessage(message);

    if (!message.read_at) {
      try {
        await API.post(`/api/message/self/${message.id}/read`);
        setMessages((prev) =>
          prev.map((item) =>
            item.id === message.id
              ? {
                  ...item,
                  read_at: new Date().toISOString(),
                }
              : item,
          ),
        );
        setSelectedMessage((prev) =>
          prev ? { ...prev, read_at: new Date().toISOString() } : prev,
        );
        refreshUnreadMessages();
      } catch (error) {
        showError(error.message);
      }
    }
  };

  const handleBatchRead = async () => {
    if (selectedRowKeys.length === 0) {
      return;
    }

    try {
      const res = await API.post('/api/message/self/read/batch', {
        ids: selectedRowKeys,
      });
      if (res.data.success) {
        const now = new Date().toISOString();
        setMessages((prev) =>
          prev.map((item) =>
            selectedRowKeys.includes(item.id)
              ? {
                  ...item,
                  read_at: item.read_at || now,
                }
              : item,
          ),
        );
        setSelectedRowKeys([]);
        refreshUnreadMessages();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const unreadCount = messages.filter((item) => !item.read_at).length;

  const columns = [
    {
      title: t('标题'),
      dataIndex: 'title',
      render: (text, record) => (
        <Space spacing={8}>
          {!record.read_at && <Tag color='red'>{t('未读')}</Tag>}
          <span className='font-medium'>{text}</span>
        </Space>
      ),
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
      title: t('来源'),
      dataIndex: 'source',
      width: 120,
      render: (value) => (
        <Tag color={value === 'system' ? 'blue' : 'green'}>
          {value === 'system' ? t('系统通知') : t('后台消息')}
        </Tag>
      ),
    },
    {
      title: t('邮件'),
      dataIndex: 'email_sent_at',
      width: 140,
      render: (value) => (
        <Tag color={value ? 'green' : 'grey'}>
          {value ? t('已发送') : t('未发送')}
        </Tag>
      ),
    },
    {
      title: t('操作'),
      width: 120,
      render: (_, record) => (
        <Button
          icon={<Eye size={14} />}
          theme='light'
          onClick={() => handleOpenMessage(record)}
        >
          {t('查看')}
        </Button>
      ),
    },
  ];

  return (
    <div className='mt-[60px] px-2'>
      <div className='mx-auto w-full max-w-6xl'>
        <div className='grid grid-cols-1 gap-4 md:grid-cols-2'>
          <Card>
            <Space spacing={12}>
              <div className='rounded-full bg-red-50 p-3 text-red-600'>
                <MailPlus size={18} />
              </div>
              <div>
                <div className='text-sm text-semi-color-text-2'>
                  {t('当前页未读')}
                </div>
                <div className='text-2xl font-semibold'>{unreadCount}</div>
              </div>
            </Space>
          </Card>

          <Card>
            <Space spacing={12}>
              <div className='rounded-full bg-blue-50 p-3 text-blue-600'>
                <MailOpen size={18} />
              </div>
              <div>
                <div className='text-sm text-semi-color-text-2'>
                  {t('消息总数')}
                </div>
                <div className='text-2xl font-semibold'>{total}</div>
              </div>
            </Space>
          </Card>
        </div>

        <Card
          style={{ marginTop: 16 }}
          title={t('我的信息')}
          headerExtraContent={
            <Space>
              <Button
                theme='light'
                disabled={selectedRowKeys.length === 0}
                onClick={handleBatchRead}
              >
                {t('批量已读')}
              </Button>
              <Button theme='light' onClick={() => loadMessages(page)}>
                {t('刷新')}
              </Button>
            </Space>
          }
        >
          <Table
            columns={columns}
            dataSource={messages}
            rowKey='id'
            loading={loading}
            rowSelection={{
              selectedRowKeys,
              onChange: (keys) => setSelectedRowKeys(keys),
              getCheckboxProps: (record) => ({
                disabled: Boolean(record.read_at),
              }),
            }}
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
                description={t('暂无信息')}
                style={{ padding: 24 }}
              />
            }
          />
        </Card>
      </div>

      <Modal
        title={selectedMessage?.title || t('信息详情')}
        visible={Boolean(selectedMessage)}
        onCancel={() => setSelectedMessage(null)}
        footer={null}
        size='large'
      >
        {selectedMessage && (
          <div className='space-y-3'>
            <Space spacing={8}>
              <Tag color={selectedMessage.read_at ? 'green' : 'red'}>
                {selectedMessage.read_at ? t('已读') : t('未读')}
              </Tag>
              <Tag color='blue'>{t('已上线')}</Tag>
            </Space>

            <Text type='secondary'>
              {selectedMessage.published_at
                ? formatDateTimeString(new Date(selectedMessage.published_at))
                : '-'}
            </Text>

            <div
              className='overflow-auto rounded-xl border border-semi-color-border bg-semi-color-bg-0'
              dangerouslySetInnerHTML={{
                __html: selectedMessage.html_content || '',
              }}
            />
          </div>
        )}
      </Modal>
    </div>
  );
};

export default Messages;
