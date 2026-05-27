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
  Modal,
  Space,
  Table,
  Tabs,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import {
  CheckCircle2,
  Download,
  FileJson,
  Plus,
  RefreshCw,
  Send,
} from 'lucide-react';
import { API, showError, showSuccess } from '../../helpers';

const PAGE_SIZE = 10;
const { Text, Title } = Typography;

const statusColor = {
  draft: 'grey',
  active: 'green',
  closed: 'blue',
  cancelled: 'red',
  pending: 'grey',
  email_sent: 'blue',
  email_failed: 'red',
  opened: 'amber',
  captured: 'green',
  migrated: 'violet',
};

const defaultExpression = 'status == "enabled" and group == "default"';

const UserMigration = () => {
  const [docs, setDocs] = useState(null);
  const [migrations, setMigrations] = useState([]);
  const [selectedMigration, setSelectedMigration] = useState(null);
  const [targets, setTargets] = useState([]);
  const [exportsData, setExportsData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [targetLoading, setTargetLoading] = useState(false);
  const [exportLoading, setExportLoading] = useState(false);
  const [preview, setPreview] = useState(null);
  const [createValue, setCreateValue] = useState({
    name: '',
    description: '',
    expression: defaultExpression,
    send_email: true,
  });
  const [appendValue, setAppendValue] = useState({
    expression: '',
    user_id: '',
    cah_id: '',
    email: '',
    send_email: true,
  });
  const [importValue, setImportValue] = useState({
    cah_id: '',
    email: '',
    data: '{\n  "account": {},\n  "records": {}\n}',
  });
  const [importResult, setImportResult] = useState(null);

  const loadDocs = async () => {
    const res = await API.get('/migrate/api/admin/expression-docs');
    if (res.data.success) {
      setDocs(res.data.data);
    }
  };

  const loadMigrations = async () => {
    setLoading(true);
    try {
      const res = await API.get('/migrate/api/admin/migrations', {
        params: { p: 1, page_size: 50 },
      });
      if (res.data.success) {
        const items = res.data.data.items || [];
        setMigrations(items);
        if (!selectedMigration && items.length > 0) {
          setSelectedMigration(items[0]);
        }
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const loadTargets = async (migration = selectedMigration) => {
    if (!migration?.migrate_id) return;
    setTargetLoading(true);
    try {
      const res = await API.get(
        `/migrate/api/admin/migrations/${migration.migrate_id}/targets`,
        { params: { p: 1, page_size: PAGE_SIZE } },
      );
      if (res.data.success) {
        setTargets(res.data.data.items || []);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setTargetLoading(false);
    }
  };

  const loadExports = async () => {
    setExportLoading(true);
    try {
      const res = await API.get('/migrate/api/admin/exports', {
        params: { limit: 20 },
      });
      if (res.data.success) {
        setExportsData(res.data.data || []);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setExportLoading(false);
    }
  };

  useEffect(() => {
    loadDocs().catch(() => null);
    loadMigrations();
    loadExports();
  }, []);

  useEffect(() => {
    loadTargets(selectedMigration);
  }, [selectedMigration?.migrate_id]);

  const previewExpression = async (expression) => {
    try {
      const res = await API.post('/migrate/api/admin/preview', {
        expression,
        limit: 10,
      });
      if (res.data.success) {
        setPreview(res.data.data);
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const createMigration = async () => {
    try {
      const res = await API.post('/migrate/api/admin/migrations', createValue);
      if (res.data.success) {
        showSuccess('迁移任务已创建');
        setPreview(null);
        await loadMigrations();
        if (res.data.data?.migration) {
          setSelectedMigration(res.data.data.migration);
        }
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const appendTargets = async () => {
    if (!selectedMigration?.migrate_id) {
      showError('请先选择迁移任务');
      return;
    }
    const payload = {
      expression: appendValue.expression,
      send_email: appendValue.send_email,
      targets: [],
    };
    if (appendValue.user_id || appendValue.cah_id) {
      payload.targets.push({
        user_id: appendValue.user_id ? Number(appendValue.user_id) : 0,
        cah_id: appendValue.cah_id,
        email: appendValue.email,
      });
    }
    try {
      const res = await API.post(
        `/migrate/api/admin/migrations/${selectedMigration.migrate_id}/targets`,
        payload,
      );
      if (res.data.success) {
        const data = res.data.data;
        showSuccess(
          `已追加 ${data.created || 0} 个用户，邮件成功 ${data.email_sent || 0}，失败 ${data.email_failed || 0}`,
        );
        await loadTargets();
        await loadMigrations();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const updateMigrationStatus = async (status) => {
    if (!selectedMigration?.migrate_id) return;
    try {
      const res = await API.put(
        `/migrate/api/admin/migrations/${selectedMigration.migrate_id}/status`,
        { status },
      );
      if (res.data.success) {
        setSelectedMigration(res.data.data);
        showSuccess('状态已更新');
        await loadMigrations();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const confirmExported = async (targetId) => {
    Modal.confirm({
      title: '确认接收方已成功导入？',
      content: '确认后原账户会以“已迁移”为原因禁用，并立即失效现有 Web 会话。',
      okText: '确认已迁移',
      cancelText: '取消',
      onOk: async () => {
        const res = await API.post('/migrate/api/admin/exports/confirm', {
          target_ids: [targetId],
        });
        if (res.data.success) {
          showSuccess('已标记为已迁移');
          await loadExports();
          await loadTargets();
          await loadMigrations();
        } else {
          showError(res.data.message);
        }
      },
    });
  };

  const importUser = async () => {
    let parsed;
    try {
      parsed = JSON.parse(importValue.data || '{}');
    } catch {
      showError('data 必须是合法 JSON');
      return;
    }
    try {
      const res = await API.post('/migrate/api/admin/imports/users', {
        cah_id: importValue.cah_id,
        email: importValue.email,
        data: parsed,
      });
      if (res.data.success) {
        const result = res.data.data;
        setImportResult(result);
        if (result?.status === 'email_failed') {
          showError(
            '用户已导入，但密码设置邮件发送失败，请使用返回的设置链接恢复处理',
          );
        } else {
          showSuccess('用户已导入，并已发送密码设置邮件');
          setImportValue({
            cah_id: '',
            email: '',
            data: '{\n  "account": {},\n  "records": {}\n}',
          });
        }
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error.message);
    }
  };

  const migrationColumns = useMemo(
    () => [
      { title: '迁移 ID', dataIndex: 'migrate_id', width: 210 },
      { title: '名称', dataIndex: 'name' },
      {
        title: '状态',
        dataIndex: 'status',
        render: (status) => <Tag color={statusColor[status]}>{status}</Tag>,
      },
      { title: '目标', dataIndex: 'target_count' },
      { title: '已发送', dataIndex: 'email_sent_count' },
      { title: '已记录', dataIndex: 'captured_count' },
      { title: '已迁移', dataIndex: 'migrated_count' },
    ],
    [],
  );

  const targetColumns = useMemo(
    () => [
      { title: 'ID', dataIndex: 'id', width: 80 },
      { title: 'CAH', dataIndex: 'cah_id', width: 120 },
      { title: '邮箱', dataIndex: 'email' },
      {
        title: '状态',
        dataIndex: 'status',
        render: (status) => <Tag color={statusColor[status]}>{status}</Tag>,
      },
      { title: '邮件错误', dataIndex: 'email_error' },
    ],
    [],
  );

  const exportColumns = useMemo(
    () => [
      { title: '目标 ID', dataIndex: 'target_id', width: 90 },
      { title: '迁移 ID', dataIndex: 'migrate_id', width: 210 },
      { title: 'CAH', dataIndex: 'cah_id', width: 120 },
      { title: '邮箱', dataIndex: 'email' },
      {
        title: '数据大小',
        dataIndex: 'data_size',
        width: 110,
        render: (value) => `${value || 0} B`,
      },
      {
        title: '已导出',
        dataIndex: 'last_exported_at',
        width: 110,
        render: (value) => (value ? '是' : '否'),
      },
      {
        title: '操作',
        render: (_, record) => (
          <Space>
            <Button
              icon={<Download size={16} />}
              onClick={() => window.open(record.download_url, '_blank')}
            >
              JSON
            </Button>
            <Button
              type='primary'
              icon={<CheckCircle2 size={16} />}
              onClick={() => confirmExported(record.target_id)}
            >
              确认已迁移
            </Button>
          </Space>
        ),
      },
    ],
    [],
  );

  return (
    <div className='mt-[60px] px-2 pb-8'>
      <div className='mx-auto max-w-[1280px]'>
        <div className='mb-4 flex flex-col gap-2 md:flex-row md:items-center md:justify-between'>
          <div>
            <Title heading={3}>用户迁移管理</Title>
            <Text type='secondary'>
              创建迁移、发送迁移指引、导出确认后的用户数据，并导入外部迁移数据。
            </Text>
          </div>
          <Button
            icon={<RefreshCw size={16} />}
            onClick={() => {
              loadMigrations();
              loadTargets();
              loadExports();
            }}
          >
            刷新
          </Button>
        </div>

        <Tabs type='line'>
          <Tabs.TabPane tab='迁移任务' itemKey='tasks'>
            <div className='grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1.1fr)_minmax(380px,0.9fr)]'>
              <Card title='任务列表'>
                <Table
                  rowKey='migrate_id'
                  columns={migrationColumns}
                  dataSource={migrations}
                  loading={loading}
                  pagination={false}
                  onRow={(record) => ({
                    onClick: () => setSelectedMigration(record),
                    style: { cursor: 'pointer' },
                  })}
                />
              </Card>

              <Card title='创建迁移'>
                <Form labelPosition='top'>
                  <Form.Input
                    label='名称'
                    value={createValue.name}
                    onChange={(value) =>
                      setCreateValue((prev) => ({ ...prev, name: value }))
                    }
                  />
                  <Form.TextArea
                    label='说明'
                    rows={2}
                    value={createValue.description}
                    onChange={(value) =>
                      setCreateValue((prev) => ({
                        ...prev,
                        description: value,
                      }))
                    }
                  />
                  <Form.TextArea
                    label='筛选表达式'
                    rows={5}
                    value={createValue.expression}
                    onChange={(value) =>
                      setCreateValue((prev) => ({
                        ...prev,
                        expression: value,
                      }))
                    }
                  />
                  <Form.Checkbox
                    checked={createValue.send_email}
                    onChange={(event) =>
                      setCreateValue((prev) => ({
                        ...prev,
                        send_email: event.target.checked,
                      }))
                    }
                  >
                    创建后立即按 Postmark 批量发送迁移邮件
                  </Form.Checkbox>
                  <Space>
                    <Button
                      icon={<FileJson size={16} />}
                      onClick={() => previewExpression(createValue.expression)}
                    >
                      预览
                    </Button>
                    <Button
                      type='primary'
                      icon={<Plus size={16} />}
                      onClick={createMigration}
                    >
                      创建迁移
                    </Button>
                  </Space>
                </Form>
                {preview && (
                  <Banner
                    className='mt-4'
                    type='info'
                    title={`命中 ${preview.count} 个用户`}
                    description={(preview.users || [])
                      .map((user) => `${user.cah_id} ${user.email}`)
                      .join('；')}
                  />
                )}
              </Card>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane tab='追加与目标' itemKey='targets'>
            <div className='grid grid-cols-1 gap-4 xl:grid-cols-[380px_minmax(0,1fr)]'>
              <Card title='当前迁移'>
                {selectedMigration ? (
                  <Space vertical align='start' className='w-full'>
                    <Text strong>{selectedMigration.name}</Text>
                    <Text copyable>{selectedMigration.migrate_id}</Text>
                    <Tag color={statusColor[selectedMigration.status]}>
                      {selectedMigration.status}
                    </Tag>
                    <Space>
                      <Button onClick={() => updateMigrationStatus('active')}>
                        启用
                      </Button>
                      <Button onClick={() => updateMigrationStatus('closed')}>
                        关闭
                      </Button>
                      <Button
                        type='danger'
                        onClick={() => updateMigrationStatus('cancelled')}
                      >
                        取消
                      </Button>
                    </Space>
                  </Space>
                ) : (
                  <Text type='secondary'>请选择迁移任务。</Text>
                )}
              </Card>

              <Card title='追加目标用户'>
                <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
                  <div>
                    <Text strong>按表达式追加</Text>
                    <TextArea
                      className='mt-2'
                      rows={5}
                      value={appendValue.expression}
                      onChange={(value) =>
                        setAppendValue((prev) => ({
                          ...prev,
                          expression: value,
                        }))
                      }
                      placeholder='email matches ".*@example\\.com$"'
                    />
                  </div>
                  <div>
                    <Text strong>按用户追加</Text>
                    <div className='mt-2 grid grid-cols-1 gap-2 md:grid-cols-3'>
                      <Input
                        placeholder='user_id'
                        value={appendValue.user_id}
                        onChange={(value) =>
                          setAppendValue((prev) => ({
                            ...prev,
                            user_id: value,
                          }))
                        }
                      />
                      <Input
                        placeholder='CAH'
                        value={appendValue.cah_id}
                        onChange={(value) =>
                          setAppendValue((prev) => ({
                            ...prev,
                            cah_id: value,
                          }))
                        }
                      />
                      <Input
                        placeholder='发送邮箱'
                        value={appendValue.email}
                        onChange={(value) =>
                          setAppendValue((prev) => ({ ...prev, email: value }))
                        }
                      />
                    </div>
                  </div>
                </div>
                <div className='mt-4'>
                  <Form.Checkbox
                    checked={appendValue.send_email}
                    onChange={(event) =>
                      setAppendValue((prev) => ({
                        ...prev,
                        send_email: event.target.checked,
                      }))
                    }
                  >
                    追加后立即发送迁移邮件
                  </Form.Checkbox>
                  <Button
                    type='primary'
                    icon={<Send size={16} />}
                    onClick={appendTargets}
                  >
                    追加目标
                  </Button>
                </div>
              </Card>

              <Card title='目标列表' className='xl:col-span-2'>
                <Table
                  rowKey='id'
                  columns={targetColumns}
                  dataSource={targets}
                  loading={targetLoading}
                  pagination={false}
                />
              </Card>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane tab='导出确认' itemKey='exports'>
            <Card
              title='待接收方确认的导出数据'
              headerExtraContent={
                <Button icon={<RefreshCw size={16} />} onClick={loadExports}>
                  刷新
                </Button>
              }
            >
              <Table
                rowKey='target_id'
                columns={exportColumns}
                dataSource={exportsData}
                loading={exportLoading}
                pagination={false}
              />
            </Card>
          </Tabs.TabPane>

          <Tabs.TabPane tab='导入用户' itemKey='imports'>
            <Card title='通过迁移 JSON 新增用户'>
              <div className='grid grid-cols-1 gap-4 lg:grid-cols-[320px_minmax(0,1fr)]'>
                <div className='space-y-3'>
                  <Input
                    placeholder='CAH'
                    value={importValue.cah_id}
                    onChange={(value) =>
                      setImportValue((prev) => ({ ...prev, cah_id: value }))
                    }
                  />
                  <Input
                    placeholder='邮箱'
                    value={importValue.email}
                    onChange={(value) =>
                      setImportValue((prev) => ({ ...prev, email: value }))
                    }
                  />
                  <Button
                    type='primary'
                    icon={<CheckCircle2 size={16} />}
                    onClick={importUser}
                  >
                    导入并发送设置密码邮件
                  </Button>
                  {importResult && (
                    <Banner
                      type={
                        importResult.status === 'email_failed'
                          ? 'warning'
                          : 'success'
                      }
                      description={
                        <div className='space-y-1'>
                          <div>
                            状态：<Tag>{importResult.status}</Tag>
                          </div>
                          {importResult.link && (
                            <Text copyable={{ content: importResult.link }}>
                              {importResult.link}
                            </Text>
                          )}
                        </div>
                      }
                    />
                  )}
                </div>
                <TextArea
                  rows={16}
                  value={importValue.data}
                  onChange={(value) =>
                    setImportValue((prev) => ({ ...prev, data: value }))
                  }
                />
              </div>
            </Card>
          </Tabs.TabPane>

          <Tabs.TabPane tab='表达式文档' itemKey='docs'>
            <Card title='CF 风格迁移筛选表达式'>
              {docs ? (
                <div className='space-y-4'>
                  <div>
                    <Text strong>语法</Text>
                    <ul className='mt-2 list-disc pl-6'>
                      {(docs.syntax || []).map((item) => (
                        <li key={item}>
                          <code>{item}</code>
                        </li>
                      ))}
                    </ul>
                  </div>
                  <div>
                    <Text strong>字段</Text>
                    <Table
                      className='mt-2'
                      rowKey='name'
                      pagination={false}
                      dataSource={docs.fields || []}
                      columns={[
                        { title: '字段', dataIndex: 'name', width: 120 },
                        { title: '说明', dataIndex: 'description' },
                        {
                          title: '运算符',
                          dataIndex: 'operators',
                          render: (ops) => (ops || []).join(', '),
                        },
                      ]}
                    />
                  </div>
                  <div>
                    <Text strong>示例</Text>
                    <div className='mt-2 flex flex-col gap-2'>
                      {(docs.examples || []).map((item) => (
                        <code key={item}>{item}</code>
                      ))}
                    </div>
                  </div>
                </div>
              ) : (
                <Text type='secondary'>文档加载中</Text>
              )}
            </Card>
          </Tabs.TabPane>
        </Tabs>
      </div>
    </div>
  );
};

export default UserMigration;
