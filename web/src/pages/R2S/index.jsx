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

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Col,
  Form,
  Modal,
  Row,
  Space,
  Spin,
  Table,
  Tabs,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  AlertTriangle,
  BarChart3,
  CheckCircle2,
  ClipboardList,
  Coins,
  CreditCard,
  Layers,
  Plus,
  ReceiptText,
  RefreshCw,
  Save,
  WalletCards,
} from 'lucide-react';
import {
  API,
  formatCurrencyAmountByCode,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';

const { Text, Title } = Typography;
const { TabPane } = Tabs;

const paymentTypes = [
  { label: '预付费', value: 'prepaid' },
  { label: '后付费', value: 'postpaid' },
  { label: '赠金', value: 'grant' },
  { label: '退款', value: 'refund' },
  { label: '调整', value: 'adjustment' },
];

const statusOptions = [
  { label: '启用', value: 'active' },
  { label: '停用', value: 'disabled' },
];

const emptySettings = {
  receipt_required: false,
  default_currency_code: 'USD',
  balance_reminder_days: 30,
  system_currency_code: 'USD',
};

const emptySummary = {
  system_currency_code: 'USD',
  recognized_revenue_amount: 0,
  recognized_cost_amount: 0,
  recognized_profit_amount: 0,
  profit_margin: 0,
  payment_system_amount: 0,
  supplier_balance_amount: 0,
  supplier_count: 0,
  active_supplier_count: 0,
  channel_binding_count: 0,
  reminder_due_count: 0,
};

const statusTag = (status, t) => {
  if (status === 'active') {
    return <Tag color='green'>{t('启用')}</Tag>;
  }
  return <Tag color='grey'>{t('停用')}</Tag>;
};

const formatAmount = (value, currencyCode) => {
  return formatCurrencyAmountByCode(Number(value || 0), currencyCode || 'USD');
};

const formatPercent = (value) => `${Number(value || 0).toFixed(2)}%`;

const renderTime = (value, t) => {
  if (!value) return t('未设置');
  return timestamp2string(value);
};

const toNumber = (value) => {
  if (value === undefined || value === null || value === '') return undefined;
  return Number(value);
};

const defaultTablePages = {
  suppliers: { page: 1, pageSize: 10, total: 0 },
  bindings: { page: 1, pageSize: 10, total: 0 },
  payments: { page: 1, pageSize: 10, total: 0 },
  balances: { page: 1, pageSize: 10, total: 0 },
  recognition: { page: 1, pageSize: 10, total: 0 },
};

const tableEndpoints = {
  suppliers: '/api/r2s/suppliers',
  bindings: '/api/r2s/channel-bindings',
  payments: '/api/r2s/payments',
  balances: '/api/r2s/balance-updates',
  recognition: '/api/r2s/recognition-records',
};

const buildQueryPath = (path, params) => {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      search.set(key, String(value));
    }
  });
  return `${path}?${search.toString()}`;
};

const R2S = () => {
  const { t } = useTranslation();
  const formApiRef = useRef(null);
  const settingsFormRef = useRef(null);
  const [activeTab, setActiveTab] = useState('dashboard');
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [modalType, setModalType] = useState(null);
  const [editingRecord, setEditingRecord] = useState(null);
  const [settings, setSettings] = useState(emptySettings);
  const [summary, setSummary] = useState(emptySummary);
  const [suppliers, setSuppliers] = useState([]);
  const [bindings, setBindings] = useState([]);
  const [payments, setPayments] = useState([]);
  const [balanceUpdates, setBalanceUpdates] = useState([]);
  const [recognitions, setRecognitions] = useState([]);
  const [promotionRows, setPromotionRows] = useState([]);
  const [channels, setChannels] = useState([]);
  const [optionSuppliers, setOptionSuppliers] = useState([]);
  const [tablePages, setTablePages] = useState(defaultTablePages);

  const supplierOptions = useMemo(
    () =>
      optionSuppliers
        .filter((supplier) => supplier.status !== 'disabled')
        .map((supplier) => ({
          label: `${supplier.name} #${supplier.id}`,
          value: supplier.id,
        })),
    [optionSuppliers],
  );

  const channelOptions = useMemo(
    () =>
      channels.map((channel) => ({
        label: `${channel.name || channel.id} #${channel.id}`,
        value: channel.id,
      })),
    [channels],
  );

  const fetchPagedList = async (path, page, pageSize, extraParams = {}) => {
    const res = await API.get(
      buildQueryPath(path, {
        ...extraParams,
        p: page,
        page_size: pageSize,
      }),
    );
    const { success, message, data } = res.data;
    if (!success) {
      throw new Error(message || t('加载失败'));
    }
    return data || { items: [], page, page_size: pageSize, total: 0 };
  };

  const loadAllOptions = async (path, extraParams = {}) => {
    const rows = [];
    let page = 1;
    let total = 0;
    do {
      const data = await fetchPagedList(path, page, 100, extraParams);
      rows.push(...(data.items || []));
      total = Number(data.total || rows.length);
      page += 1;
    } while (rows.length < total && page <= 50);
    return rows;
  };

  const setTableRows = (key, rows) => {
    if (key === 'suppliers') setSuppliers(rows);
    if (key === 'bindings') setBindings(rows);
    if (key === 'payments') setPayments(rows);
    if (key === 'balances') setBalanceUpdates(rows);
    if (key === 'recognition') setRecognitions(rows);
  };

  const applyTablePage = (key, data, page, pageSize) => {
    setTableRows(key, data.items || []);
    setTablePages((prev) => ({
      ...prev,
      [key]: {
        page: Number(data.page || page),
        pageSize: Number(data.page_size || pageSize),
        total: Number(data.total || 0),
      },
    }));
  };

  const loadTablePage = async (key, page, pageSize) => {
    setLoading(true);
    try {
      const data = await fetchPagedList(tableEndpoints[key], page, pageSize);
      applyTablePage(key, data, page, pageSize);
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const refresh = async ({ resetLists = false } = {}) => {
    setLoading(true);
    try {
      const pageConfig = resetLists ? defaultTablePages : tablePages;
      const tableRequests = Object.entries(tableEndpoints).map(
        async ([key, path]) => {
          const meta = pageConfig[key] || defaultTablePages[key];
          const data = await fetchPagedList(path, meta.page, meta.pageSize);
          return { key, data, page: meta.page, pageSize: meta.pageSize };
        },
      );
      const [
        settingsRes,
        summaryRes,
        promotionData,
        channelData,
        supplierOptionData,
        ...tableResults
      ] = await Promise.all([
        API.get('/api/r2s/settings'),
        API.get('/api/r2s/summary'),
        API.get('/api/r2s/promotion-profitability'),
        loadAllOptions('/api/channel/'),
        loadAllOptions('/api/r2s/suppliers', { status: 'active' }),
        ...tableRequests,
      ]);

      if (settingsRes.data.success) {
        setSettings(settingsRes.data.data || emptySettings);
        settingsFormRef.current?.setValues(
          settingsRes.data.data || emptySettings,
        );
      }
      if (summaryRes.data.success) {
        setSummary(summaryRes.data.data || emptySummary);
      }
      if (promotionData.data.success) {
        setPromotionRows(promotionData.data.data || []);
      }
      setChannels(channelData);
      setOptionSuppliers(supplierOptionData);
      tableResults.forEach(({ key, data, page, pageSize }) => {
        applyTablePage(key, data, page, pageSize);
      });
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  const closeModal = () => {
    setModalType(null);
    setEditingRecord(null);
    formApiRef.current?.reset();
  };

  const openModal = (type, record = null) => {
    setModalType(type);
    setEditingRecord(record);
    setTimeout(() => {
      formApiRef.current?.setValues(getInitialValues(type, record));
    }, 0);
  };

  const getInitialValues = (type, record) => {
    if (record) {
      return {
        ...record,
        receipt_url: record.receipt_url || '',
        note: record.note || '',
      };
    }
    switch (type) {
      case 'supplier':
        return {
          status: 'active',
          default_currency_code: settings.default_currency_code,
          default_exchange_rate: 1,
          balance_amount: 0,
          balance_currency_code: settings.default_currency_code,
          balance_reminder_days: settings.balance_reminder_days,
        };
      case 'binding':
        return { status: 'active', group_multiplier: 1 };
      case 'payment':
        return {
          payment_type: 'prepaid',
          currency_code: settings.default_currency_code,
          exchange_rate: 1,
        };
      case 'balance':
        return {
          currency_code: settings.default_currency_code,
          exchange_rate: 1,
        };
      default:
        return {};
    }
  };

  const saveSettings = async (values) => {
    setSaving(true);
    try {
      const res = await API.put('/api/r2s/settings', {
        receipt_required: Boolean(values.receipt_required),
        default_currency_code: (values.default_currency_code || '').trim(),
        balance_reminder_days: Number(values.balance_reminder_days || 0),
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('保存成功'));
        setSettings(data || emptySettings);
        await refresh();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSaving(false);
    }
  };

  const submitModal = async (values) => {
    setSaving(true);
    try {
      let res;
      if (modalType === 'supplier') {
        const payload = {
          name: values.name,
          description: values.description || '',
          status: values.status || 'active',
          default_currency_code: values.default_currency_code,
          default_exchange_rate: Number(values.default_exchange_rate || 1),
          balance_reminder_days: Number(values.balance_reminder_days || 0),
        };
        if (!editingRecord?.id) {
          payload.balance_amount = Number(values.balance_amount || 0);
          payload.balance_currency_code = values.balance_currency_code;
        }
        res = editingRecord?.id
          ? await API.put(`/api/r2s/suppliers/${editingRecord.id}`, payload)
          : await API.post('/api/r2s/suppliers', payload);
      } else if (modalType === 'binding') {
        const payload = {
          supplier_id: Number(values.supplier_id),
          channel_id: Number(values.channel_id),
          upstream_group_name: values.upstream_group_name || '',
          group_multiplier: Number(values.group_multiplier || 1),
          status: values.status || 'active',
        };
        res = editingRecord?.id
          ? await API.put(
              `/api/r2s/channel-bindings/${editingRecord.id}`,
              payload,
            )
          : await API.post('/api/r2s/channel-bindings', payload);
      } else if (modalType === 'payment') {
        res = await API.post('/api/r2s/payments', {
          supplier_id: Number(values.supplier_id),
          payment_type: values.payment_type,
          amount: Number(values.amount),
          currency_code: values.currency_code,
          exchange_rate: Number(values.exchange_rate || 1),
          balance_after: toNumber(values.balance_after),
          receipt_url: values.receipt_url || '',
          note: values.note || '',
        });
      } else if (modalType === 'balance') {
        res = await API.post('/api/r2s/balance-updates', {
          supplier_id: Number(values.supplier_id),
          balance_after: Number(values.balance_after),
          currency_code: values.currency_code,
          exchange_rate: Number(values.exchange_rate || 1),
          balance_reminder_days: toNumber(values.balance_reminder_days),
          note: values.note || '',
        });
      }

      const { success, message } = res.data;
      if (success) {
        showSuccess(t('保存成功'));
        closeModal();
        await refresh({ resetLists: true });
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setSaving(false);
    }
  };

  const disableRecord = (type, id) => {
    Modal.confirm({
      title: t('确认停用'),
      content: t('停用后历史记录仍会保留。'),
      onOk: async () => {
        try {
          const path =
            type === 'supplier'
              ? `/api/r2s/suppliers/${id}`
              : `/api/r2s/channel-bindings/${id}`;
          const res = await API.delete(path);
          const { success, message } = res.data;
          if (success) {
            showSuccess(t('操作成功'));
            await refresh();
          } else {
            showError(message);
          }
        } catch (error) {
          showError(error.message);
        }
      },
    });
  };

  const updateSupplierStatus = (record, status) => {
    const isEnable = status === 'active';
    Modal.confirm({
      title: isEnable ? t('确认启用供应商') : t('确认停用供应商'),
      content: isEnable
        ? t('启用后该供应商会重新进入余额统计和提醒。')
        : t('停用后历史记录仍会保留，并且不再进入余额统计和提醒。'),
      onOk: async () => {
        try {
          const res = isEnable
            ? await API.post(`/api/r2s/suppliers/${record.id}/enable`)
            : await API.delete(`/api/r2s/suppliers/${record.id}`);
          const { success, message } = res.data;
          if (success) {
            showSuccess(t('操作成功'));
            await refresh({ resetLists: true });
          } else {
            showError(message);
          }
        } catch (error) {
          showError(error.message);
        }
      },
    });
  };

  const deleteSupplier = (record) => {
    Modal.confirm({
      title: t('确认删除供应商'),
      content: t(
        '仅未产生付款、余额更新或识别记录的供应商可以删除；关联渠道绑定会一并删除。',
      ),
      onOk: async () => {
        try {
          const res = await API.delete(
            `/api/r2s/suppliers/${record.id}/permanent`,
          );
          const { success, message } = res.data;
          if (success) {
            showSuccess(t('删除成功'));
            await refresh({ resetLists: true });
          } else {
            showError(message);
          }
        } catch (error) {
          showError(error.message);
        }
      },
    });
  };

  const supplierColumns = [
    { title: t('ID'), dataIndex: 'id', width: 80 },
    { title: t('供应商名称'), dataIndex: 'name', width: 180 },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 100,
      render: (status) => statusTag(status, t),
    },
    {
      title: t('余额'),
      dataIndex: 'balance_amount',
      width: 160,
      render: (_, record) =>
        formatAmount(record.balance_amount, record.balance_currency_code),
    },
    {
      title: t('系统折算'),
      dataIndex: 'system_balance_amount',
      width: 160,
      render: (_, record) =>
        formatAmount(
          record.system_balance_amount,
          summary.system_currency_code,
        ),
    },
    { title: t('默认汇率'), dataIndex: 'default_exchange_rate', width: 120 },
    { title: t('渠道数'), dataIndex: 'channel_count', width: 100 },
    {
      title: t('下次提醒'),
      dataIndex: 'next_balance_reminder_at',
      width: 180,
      render: (value) => renderTime(value, t),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      width: 240,
      fixed: 'right',
      render: (_, record) => (
        <Space wrap>
          <Button size='small' onClick={() => openModal('supplier', record)}>
            {t('编辑')}
          </Button>
          {record.status === 'disabled' ? (
            <Button
              size='small'
              type='primary'
              onClick={() => updateSupplierStatus(record, 'active')}
            >
              {t('启用')}
            </Button>
          ) : (
            <Button
              size='small'
              type='danger'
              onClick={() => updateSupplierStatus(record, 'disabled')}
            >
              {t('停用')}
            </Button>
          )}
          <Button
            size='small'
            type='danger'
            onClick={() => deleteSupplier(record)}
          >
            {t('删除')}
          </Button>
        </Space>
      ),
    },
  ];

  const bindingColumns = [
    { title: t('ID'), dataIndex: 'id', width: 80 },
    { title: t('供应商'), dataIndex: 'supplier_name', width: 180 },
    { title: t('渠道'), dataIndex: 'channel_name_snapshot', width: 180 },
    { title: t('上游分组'), dataIndex: 'upstream_group_name', width: 140 },
    { title: t('分组倍率'), dataIndex: 'group_multiplier', width: 120 },
    {
      title: t('状态'),
      dataIndex: 'status',
      width: 100,
      render: (status) => statusTag(status, t),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      width: 150,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Button size='small' onClick={() => openModal('binding', record)}>
            {t('编辑')}
          </Button>
          <Button
            size='small'
            type='danger'
            onClick={() => disableRecord('binding', record.id)}
          >
            {t('停用')}
          </Button>
        </Space>
      ),
    },
  ];

  const paymentColumns = [
    { title: t('ID'), dataIndex: 'id', width: 80 },
    { title: t('供应商'), dataIndex: 'supplier_name_snapshot', width: 180 },
    { title: t('类型'), dataIndex: 'payment_type', width: 110 },
    {
      title: t('金额'),
      dataIndex: 'amount',
      width: 160,
      render: (_, record) => formatAmount(record.amount, record.currency_code),
    },
    { title: t('汇率'), dataIndex: 'exchange_rate', width: 100 },
    {
      title: t('余额变化'),
      dataIndex: 'balance_after',
      width: 170,
      render: (_, record) =>
        `${record.balance_before} → ${record.balance_after}`,
    },
    {
      title: t('收据'),
      dataIndex: 'receipt_url',
      width: 120,
      render: (value) =>
        value ? (
          <Typography.Text link={{ href: value, target: '_blank' }}>
            {t('查看')}
          </Typography.Text>
        ) : (
          t('未上传')
        ),
    },
    {
      title: t('付款时间'),
      dataIndex: 'paid_at',
      width: 180,
      render: (value) => renderTime(value, t),
    },
  ];

  const balanceColumns = [
    { title: t('ID'), dataIndex: 'id', width: 80 },
    { title: t('供应商'), dataIndex: 'supplier_name_snapshot', width: 180 },
    { title: t('类型'), dataIndex: 'update_type', width: 110 },
    {
      title: t('余额变化'),
      dataIndex: 'balance_after',
      width: 170,
      render: (_, record) =>
        `${record.balance_before} → ${record.balance_after}`,
    },
    {
      title: t('折算变化'),
      dataIndex: 'system_delta_amount',
      width: 160,
      render: (_, record) =>
        formatAmount(record.system_delta_amount, summary.system_currency_code),
    },
    {
      title: t('下次提醒'),
      dataIndex: 'next_reminder_at',
      width: 180,
      render: (value) => renderTime(value, t),
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      width: 180,
      render: (value) => renderTime(value, t),
    },
  ];

  const recognitionColumns = [
    { title: t('ID'), dataIndex: 'id', width: 80 },
    { title: t('供应商'), dataIndex: 'supplier_name_snapshot', width: 180 },
    { title: t('渠道'), dataIndex: 'channel_name_snapshot', width: 180 },
    { title: t('来源'), dataIndex: 'source_type', width: 110 },
    {
      title: t('收入'),
      dataIndex: 'system_revenue_amount',
      width: 150,
      render: (_, record) =>
        formatAmount(
          record.system_revenue_amount,
          summary.system_currency_code,
        ),
    },
    {
      title: t('成本'),
      dataIndex: 'system_cost_amount',
      width: 150,
      render: (_, record) =>
        formatAmount(record.system_cost_amount, summary.system_currency_code),
    },
    {
      title: t('利润率'),
      dataIndex: 'profit_margin',
      width: 110,
      render: (value) => formatPercent(value),
    },
    {
      title: t('倍率快照'),
      dataIndex: 'group_multiplier_snapshot',
      width: 120,
    },
    {
      title: t('周期'),
      dataIndex: 'period_start',
      width: 260,
      render: (_, record) =>
        `${renderTime(record.period_start, t)} - ${renderTime(record.period_end, t)}`,
    },
  ];

  const promotionColumns = [
    { title: t('活动 ID'), dataIndex: 'campaign_id', width: 100 },
    { title: t('活动名称'), dataIndex: 'campaign_name', width: 200 },
    { title: t('核销次数'), dataIndex: 'top_up_count', width: 100 },
    {
      title: t('实收'),
      dataIndex: 'net_revenue_amount',
      width: 150,
      render: (_, record) =>
        formatAmount(record.net_revenue_amount, record.currency_code),
    },
    {
      title: t('已识别成本'),
      dataIndex: 'recognized_cost_amount',
      width: 150,
      render: (_, record) =>
        formatAmount(
          record.recognized_cost_amount,
          record.system_currency_code || summary.system_currency_code,
        ),
    },
    {
      title: t('利润'),
      dataIndex: 'profit_amount',
      width: 150,
      render: (_, record) =>
        record.profit_calculated === false ? (
          <Tag color='orange'>{t('需汇率')}</Tag>
        ) : (
          formatAmount(
            record.profit_amount,
            record.system_currency_code || summary.system_currency_code,
          )
        ),
    },
    {
      title: t('利润率'),
      dataIndex: 'profit_margin',
      width: 110,
      render: (value, record) =>
        record.profit_calculated === false ? '-' : formatPercent(value),
    },
  ];

  const modalTitle = {
    supplier: editingRecord ? t('编辑供应商') : t('新增供应商'),
    binding: editingRecord ? t('编辑渠道绑定') : t('新增渠道绑定'),
    payment: t('记录付款'),
    balance: t('更新供应商余额'),
  }[modalType];

  const dueSuppliers = optionSuppliers.filter((supplier) => {
    const nextReminderAt = Number(supplier.next_balance_reminder_at || 0);
    return (
      supplier.status !== 'disabled' &&
      nextReminderAt > 0 &&
      nextReminderAt <= Math.floor(Date.now() / 1000)
    );
  });

  const openBalanceForSupplier = (supplier) => {
    setModalType('balance');
    setEditingRecord(null);
    setTimeout(() => {
      formApiRef.current?.setValues({
        supplier_id: supplier.id,
        balance_after: Number(supplier.balance_amount || 0),
        currency_code:
          supplier.balance_currency_code ||
          supplier.default_currency_code ||
          settings.default_currency_code,
        exchange_rate: Number(supplier.default_exchange_rate || 1),
        balance_reminder_days:
          supplier.balance_reminder_days ?? settings.balance_reminder_days,
      });
    }, 0);
  };

  return (
    <div className='mt-[60px] px-2 pb-6'>
      <Spin spinning={loading} size='large'>
        <div className='mb-4 flex flex-col md:flex-row md:items-center md:justify-between gap-3'>
          <div>
            <Title heading={3} className='m-0'>
              R2S
            </Title>
            <Text type='secondary'>
              {t('收入识别随看板刷新同步汇总，优先处理提醒和配置缺口')}
            </Text>
          </div>
          <Space wrap>
            <Button icon={<RefreshCw size={15} />} onClick={refresh}>
              {t('刷新')}
            </Button>
            <Button
              type='primary'
              icon={<Plus size={15} />}
              onClick={() => openModal('supplier')}
            >
              {t('新增供应商')}
            </Button>
          </Space>
        </div>

        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          type='card'
          className='mb-3'
        >
          <TabPane tab={t('仪表板')} itemKey='dashboard' />
          <TabPane tab={t('供应商')} itemKey='suppliers' />
          <TabPane tab={t('渠道绑定')} itemKey='bindings' />
          <TabPane tab={t('付款历史')} itemKey='payments' />
          <TabPane tab={t('余额更新')} itemKey='balances' />
          <TabPane tab={t('识别明细')} itemKey='recognition' />
          <TabPane tab={t('促销盈利性')} itemKey='promotions' />
        </Tabs>

        {activeTab === 'dashboard' ? (
          <R2SDashboard
            t={t}
            summary={summary}
            settings={settings}
            settingsFormRef={settingsFormRef}
            saving={saving}
            dueSuppliers={dueSuppliers}
            onSaveSettings={saveSettings}
            onOpenModal={openModal}
            onOpenBalanceForSupplier={openBalanceForSupplier}
            onOpenTab={setActiveTab}
          />
        ) : (
          <Card
            title={renderTabTitle(activeTab, t)}
            headerExtraContent={renderTabAction(activeTab, openModal, t)}
          >
            {activeTab === 'suppliers' && (
              <R2STable
                columns={supplierColumns}
                data={suppliers}
                pageInfo={tablePages.suppliers}
                onPageChange={(page) =>
                  loadTablePage(
                    'suppliers',
                    page,
                    tablePages.suppliers.pageSize,
                  )
                }
                onPageSizeChange={(pageSize) =>
                  loadTablePage('suppliers', 1, pageSize)
                }
              />
            )}
            {activeTab === 'bindings' && (
              <R2STable
                columns={bindingColumns}
                data={bindings}
                pageInfo={tablePages.bindings}
                onPageChange={(page) =>
                  loadTablePage('bindings', page, tablePages.bindings.pageSize)
                }
                onPageSizeChange={(pageSize) =>
                  loadTablePage('bindings', 1, pageSize)
                }
              />
            )}
            {activeTab === 'payments' && (
              <R2STable
                columns={paymentColumns}
                data={payments}
                pageInfo={tablePages.payments}
                onPageChange={(page) =>
                  loadTablePage('payments', page, tablePages.payments.pageSize)
                }
                onPageSizeChange={(pageSize) =>
                  loadTablePage('payments', 1, pageSize)
                }
              />
            )}
            {activeTab === 'balances' && (
              <R2STable
                columns={balanceColumns}
                data={balanceUpdates}
                pageInfo={tablePages.balances}
                onPageChange={(page) =>
                  loadTablePage('balances', page, tablePages.balances.pageSize)
                }
                onPageSizeChange={(pageSize) =>
                  loadTablePage('balances', 1, pageSize)
                }
              />
            )}
            {activeTab === 'recognition' && (
              <R2STable
                columns={recognitionColumns}
                data={recognitions}
                pageInfo={tablePages.recognition}
                onPageChange={(page) =>
                  loadTablePage(
                    'recognition',
                    page,
                    tablePages.recognition.pageSize,
                  )
                }
                onPageSizeChange={(pageSize) =>
                  loadTablePage('recognition', 1, pageSize)
                }
              />
            )}
            {activeTab === 'promotions' && (
              <R2STable columns={promotionColumns} data={promotionRows} />
            )}
          </Card>
        )}

        <Modal
          title={modalTitle}
          visible={Boolean(modalType)}
          onCancel={closeModal}
          onOk={() => formApiRef.current?.submitForm()}
          confirmLoading={saving}
          size='large'
        >
          <Form
            getFormApi={(api) => {
              formApiRef.current = api;
            }}
            initValues={getInitialValues(modalType, editingRecord)}
            onSubmit={submitModal}
          >
            {renderModalFields({
              type: modalType,
              t,
              supplierOptions,
              channelOptions,
              settings,
              isEdit: Boolean(editingRecord?.id),
            })}
          </Form>
        </Modal>
      </Spin>
    </div>
  );
};

const MetricCard = ({ icon, title, value, extra }) => {
  return (
    <Card bodyStyle={{ padding: 16 }}>
      <Space align='start'>
        <div className='text-[var(--semi-color-primary)] mt-1'>{icon}</div>
        <div>
          <Text type='secondary'>{title}</Text>
          <div className='text-xl font-semibold leading-7'>{value}</div>
          <Text type='tertiary'>{extra}</Text>
        </div>
      </Space>
    </Card>
  );
};

const R2SDashboard = ({
  t,
  summary,
  settings,
  settingsFormRef,
  saving,
  dueSuppliers,
  onSaveSettings,
  onOpenModal,
  onOpenBalanceForSupplier,
  onOpenTab,
}) => {
  const reminderCount = Number(summary.reminder_due_count || 0);
  const activeSupplierCount = Number(summary.active_supplier_count || 0);
  const supplierCount = Number(summary.supplier_count || 0);
  const bindingCount = Number(summary.channel_binding_count || 0);
  const hasSuppliers = activeSupplierCount > 0;
  const hasBindings = bindingCount > 0;
  const hasRecognizedData =
    Number(summary.recognized_revenue_amount || 0) > 0 ||
    Number(summary.recognized_cost_amount || 0) > 0;
  const visibleDueSuppliers = dueSuppliers.slice(0, 5);
  const dueSupplierColumns = [
    { title: t('供应商'), dataIndex: 'name', width: 180 },
    {
      title: t('当前余额'),
      dataIndex: 'balance_amount',
      width: 160,
      render: (_, record) =>
        formatAmount(record.balance_amount, record.balance_currency_code),
    },
    {
      title: t('系统折算'),
      dataIndex: 'system_balance_amount',
      width: 160,
      render: (_, record) =>
        formatAmount(
          record.system_balance_amount,
          summary.system_currency_code,
        ),
    },
    {
      title: t('提醒时间'),
      dataIndex: 'next_balance_reminder_at',
      width: 180,
      render: (value) => renderTime(value, t),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      width: 120,
      render: (_, record) => (
        <Button
          size='small'
          type='primary'
          onClick={() => onOpenBalanceForSupplier(record)}
        >
          {t('更新余额')}
        </Button>
      ),
    },
  ];

  return (
    <>
      <Row gutter={[12, 12]} className='mb-3'>
        <Col xs={24} md={12} xl={6}>
          <MetricCard
            icon={<BarChart3 size={18} />}
            title={t('已识别利润')}
            value={formatAmount(
              summary.recognized_profit_amount,
              summary.system_currency_code,
            )}
            extra={formatPercent(summary.profit_margin)}
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <MetricCard
            icon={<Coins size={18} />}
            title={t('已识别收入')}
            value={formatAmount(
              summary.recognized_revenue_amount,
              summary.system_currency_code,
            )}
            extra={formatAmount(
              summary.recognized_cost_amount,
              summary.system_currency_code,
            )}
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <MetricCard
            icon={<WalletCards size={18} />}
            title={t('供应商余额')}
            value={formatAmount(
              summary.supplier_balance_amount,
              summary.system_currency_code,
            )}
            extra={`${activeSupplierCount}/${supplierCount}`}
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <MetricCard
            icon={<ReceiptText size={18} />}
            title={t('余额提醒')}
            value={String(reminderCount)}
            extra={`${bindingCount} ${t('个渠道绑定')}`}
          />
        </Col>
      </Row>

      <Row gutter={[12, 12]} className='mb-3'>
        <Col xs={24} md={12} xl={6}>
          <DashboardActionCard
            icon={
              reminderCount > 0 ? (
                <AlertTriangle size={18} />
              ) : (
                <CheckCircle2 size={18} />
              )
            }
            status={reminderCount > 0 ? 'warning' : 'ok'}
            tagText={reminderCount > 0 ? t('待处理') : t('正常')}
            title={t('余额提醒')}
            description={
              reminderCount > 0
                ? t('有供应商余额需要更新，避免利润识别继续使用旧余额。')
                : t('当前没有到期的供应商余额提醒。')
            }
            buttonText={reminderCount > 0 ? t('处理提醒') : t('查看供应商')}
            onClick={() => onOpenTab('suppliers')}
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <DashboardActionCard
            icon={
              hasSuppliers ? (
                <CheckCircle2 size={18} />
              ) : (
                <AlertTriangle size={18} />
              )
            }
            status={hasSuppliers ? 'ok' : 'warning'}
            tagText={hasSuppliers ? t('已配置') : t('缺失')}
            title={t('供应商配置')}
            description={
              hasSuppliers
                ? t('已有启用供应商，可以继续维护付款和余额。')
                : t('先新增供应商，后续付款、余额和识别才能归属。')
            }
            buttonText={hasSuppliers ? t('管理供应商') : t('新增供应商')}
            onClick={() =>
              hasSuppliers ? onOpenTab('suppliers') : onOpenModal('supplier')
            }
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <DashboardActionCard
            icon={
              hasBindings ? (
                <CheckCircle2 size={18} />
              ) : (
                <AlertTriangle size={18} />
              )
            }
            status={hasBindings ? 'ok' : 'warning'}
            tagText={hasBindings ? t('已绑定') : t('缺失')}
            title={t('渠道绑定')}
            description={
              hasBindings
                ? t('渠道已关联供应商，收入识别会按绑定快照归集。')
                : t('绑定渠道和供应商后，后续收入才能自动进入识别明细。')
            }
            buttonText={
              hasBindings
                ? t('查看绑定')
                : hasSuppliers
                  ? t('新增绑定')
                  : t('新增供应商')
            }
            onClick={() => {
              if (hasBindings) {
                onOpenTab('bindings');
                return;
              }
              onOpenModal(hasSuppliers ? 'binding' : 'supplier');
            }}
          />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <DashboardActionCard
            icon={<ClipboardList size={18} />}
            status={hasRecognizedData ? 'ok' : 'neutral'}
            tagText={t('看板汇总')}
            title={t('收入识别')}
            description={t('识别数据随刷新汇总，不需要手动创建报告。')}
            buttonText={t('查看明细')}
            onClick={() => onOpenTab('recognition')}
          />
        </Col>
      </Row>

      <Card
        className='mb-3'
        title={
          <Space>
            <AlertTriangle size={16} />
            <span>{t('待处理余额提醒')}</span>
          </Space>
        }
        headerExtraContent={
          dueSuppliers.length > 5 ? (
            <Text type='tertiary'>{t('仅显示最早需要处理的 5 个供应商')}</Text>
          ) : null
        }
      >
        {visibleDueSuppliers.length > 0 ? (
          <Table
            columns={dueSupplierColumns}
            dataSource={visibleDueSuppliers}
            rowKey='id'
            size='small'
            scroll={{ x: 'max-content' }}
            pagination={false}
          />
        ) : (
          <div className='py-6 text-center'>
            <CheckCircle2
              size={22}
              className='mx-auto mb-2 text-[var(--semi-color-success)]'
            />
            <Text type='secondary'>{t('暂无到期余额提醒')}</Text>
          </div>
        )}
      </Card>

      <Card className='mb-3' title={t('看板设置')}>
        <Form
          layout='horizontal'
          initValues={settings}
          getFormApi={(api) => {
            settingsFormRef.current = api;
          }}
          onSubmit={onSaveSettings}
        >
          <Row gutter={[12, 8]}>
            <Col xs={24} md={6}>
              <Form.Switch
                field='receipt_required'
                label={t('付款必须上传收据')}
              />
            </Col>
            <Col xs={24} md={6}>
              <Form.Input field='default_currency_code' label={t('默认货币')} />
            </Col>
            <Col xs={24} md={6}>
              <Form.InputNumber
                field='balance_reminder_days'
                label={t('默认提醒天数')}
                min={0}
              />
            </Col>
            <Col xs={24} md={6}>
              <Button
                type='primary'
                icon={<Save size={15} />}
                loading={saving}
                onClick={() => settingsFormRef.current?.submitForm()}
              >
                {t('保存设置')}
              </Button>
            </Col>
          </Row>
        </Form>
      </Card>
    </>
  );
};

const DashboardActionCard = ({
  icon,
  status,
  tagText,
  title,
  description,
  buttonText,
  onClick,
}) => {
  const iconClassName =
    status === 'warning'
      ? 'text-[var(--semi-color-warning)]'
      : 'text-[var(--semi-color-primary)]';
  const tagColor =
    status === 'warning' ? 'orange' : status === 'neutral' ? 'blue' : 'green';

  return (
    <Card bodyStyle={{ padding: 16 }}>
      <Space align='start' className='w-full'>
        <div className={`${iconClassName} mt-1`}>{icon}</div>
        <div className='min-w-0 flex-1'>
          <div className='mb-1 flex items-start justify-between gap-2'>
            <Text strong>{title}</Text>
            <Tag color={tagColor}>{tagText}</Tag>
          </div>
          <Text type='secondary'>{description}</Text>
          <div className='mt-3'>
            <Button size='small' onClick={onClick}>
              {buttonText}
            </Button>
          </div>
        </div>
      </Space>
    </Card>
  );
};

const R2STable = ({
  columns,
  data,
  pageInfo,
  onPageChange,
  onPageSizeChange,
}) => {
  const pagination = pageInfo
    ? {
        currentPage: pageInfo.page,
        pageSize: pageInfo.pageSize,
        total: pageInfo.total,
        pageSizeOpts: [10, 20, 50, 100],
        showSizeChanger: true,
        onPageChange,
        onPageSizeChange,
      }
    : { pageSize: 10, showSizeChanger: true };

  return (
    <Table
      columns={columns}
      dataSource={data}
      rowKey='id'
      size='middle'
      scroll={{ x: 'max-content' }}
      pagination={pagination}
    />
  );
};

const renderTabTitle = (activeTab, t) => {
  const map = {
    suppliers: [<WalletCards size={16} />, t('供应商')],
    bindings: [<Layers size={16} />, t('渠道绑定')],
    payments: [<CreditCard size={16} />, t('付款历史')],
    balances: [<ReceiptText size={16} />, t('余额更新')],
    recognition: [<BarChart3 size={16} />, t('识别明细')],
    promotions: [<Coins size={16} />, t('促销盈利性')],
  };
  const [icon, label] = map[activeTab] || map.suppliers;
  return (
    <Space>
      {icon}
      <span>{label}</span>
    </Space>
  );
};

const renderTabAction = (activeTab, openModal, t) => {
  const actionMap = {
    suppliers: ['supplier', t('新增供应商')],
    bindings: ['binding', t('新增绑定')],
    payments: ['payment', t('记录付款')],
    balances: ['balance', t('更新余额')],
  };
  const action = actionMap[activeTab];
  if (!action) return null;
  return (
    <Button
      type='primary'
      icon={<Plus size={15} />}
      onClick={() => openModal(action[0])}
    >
      {action[1]}
    </Button>
  );
};

const renderModalFields = ({
  type,
  t,
  supplierOptions,
  channelOptions,
  settings,
  isEdit,
}) => {
  if (type === 'supplier') {
    return (
      <Row gutter={12}>
        <Col span={12}>
          <Form.Input
            field='name'
            label={t('供应商名称')}
            rules={[{ required: true, message: t('请输入供应商名称') }]}
          />
        </Col>
        <Col span={12}>
          <Form.Select
            field='status'
            label={t('状态')}
            optionList={statusOptions.map((item) => ({
              ...item,
              label: t(item.label),
            }))}
          />
        </Col>
        <Col span={24}>
          <Form.TextArea field='description' label={t('描述')} rows={2} />
        </Col>
        <Col span={12}>
          <Form.Input field='default_currency_code' label={t('默认货币')} />
        </Col>
        <Col span={12}>
          <Form.InputNumber
            field='default_exchange_rate'
            label={t('默认汇率')}
            min={0.000001}
          />
        </Col>
        {!isEdit && (
          <>
            <Col span={12}>
              <Form.InputNumber field='balance_amount' label={t('初始余额')} />
            </Col>
            <Col span={12}>
              <Form.Input field='balance_currency_code' label={t('余额货币')} />
            </Col>
          </>
        )}
        <Col span={12}>
          <Form.InputNumber
            field='balance_reminder_days'
            label={t('提醒天数')}
            min={0}
          />
        </Col>
      </Row>
    );
  }

  if (type === 'binding') {
    return (
      <Row gutter={12}>
        <Col span={12}>
          <Form.Select
            field='supplier_id'
            label={t('供应商')}
            optionList={supplierOptions}
            rules={[{ required: true, message: t('请选择供应商') }]}
            filter
          />
        </Col>
        <Col span={12}>
          <Form.Select
            field='channel_id'
            label={t('渠道')}
            optionList={channelOptions}
            rules={[{ required: true, message: t('请选择渠道') }]}
            filter
          />
        </Col>
        <Col span={12}>
          <Form.Input field='upstream_group_name' label={t('上游分组')} />
        </Col>
        <Col span={12}>
          <Form.InputNumber
            field='group_multiplier'
            label={t('分组倍率')}
            min={0.000001}
          />
        </Col>
        <Col span={12}>
          <Form.Select
            field='status'
            label={t('状态')}
            optionList={statusOptions.map((item) => ({
              ...item,
              label: t(item.label),
            }))}
          />
        </Col>
      </Row>
    );
  }

  if (type === 'payment') {
    return (
      <Row gutter={12}>
        <Col span={12}>
          <Form.Select
            field='supplier_id'
            label={t('供应商')}
            optionList={supplierOptions}
            rules={[{ required: true, message: t('请选择供应商') }]}
            filter
          />
        </Col>
        <Col span={12}>
          <Form.Select
            field='payment_type'
            label={t('付款类型')}
            optionList={paymentTypes.map((item) => ({
              ...item,
              label: t(item.label),
            }))}
          />
        </Col>
        <Col span={8}>
          <Form.InputNumber
            field='amount'
            label={t('付款金额')}
            min={0.000001}
            rules={[{ required: true, message: t('请输入付款金额') }]}
          />
        </Col>
        <Col span={8}>
          <Form.Input field='currency_code' label={t('付款货币')} />
        </Col>
        <Col span={8}>
          <Form.InputNumber
            field='exchange_rate'
            label={t('本次汇率')}
            min={0.000001}
          />
        </Col>
        <Col span={12}>
          <Form.InputNumber
            field='balance_after'
            label={t('付款后余额')}
            placeholder={t('留空按付款类型自动计算')}
          />
        </Col>
        <Col span={12}>
          <Form.Input field='receipt_url' label={t('收据/截图地址')} />
        </Col>
        <Col span={24}>
          <Form.TextArea field='note' label={t('备注')} rows={2} />
        </Col>
      </Row>
    );
  }

  if (type === 'balance') {
    return (
      <Row gutter={12}>
        <Col span={12}>
          <Form.Select
            field='supplier_id'
            label={t('供应商')}
            optionList={supplierOptions}
            rules={[{ required: true, message: t('请选择供应商') }]}
            filter
          />
        </Col>
        <Col span={12}>
          <Form.InputNumber
            field='balance_after'
            label={t('更新后余额')}
            rules={[{ required: true, message: t('请输入更新后余额') }]}
          />
        </Col>
        <Col span={8}>
          <Form.Input
            field='currency_code'
            label={t('余额货币')}
            placeholder={settings.default_currency_code}
          />
        </Col>
        <Col span={8}>
          <Form.InputNumber
            field='exchange_rate'
            label={t('本次汇率')}
            min={0.000001}
          />
        </Col>
        <Col span={8}>
          <Form.InputNumber
            field='balance_reminder_days'
            label={t('提醒天数')}
            min={0}
            placeholder={t('留空不变')}
          />
        </Col>
        <Col span={24}>
          <Form.TextArea field='note' label={t('备注')} rows={2} />
        </Col>
      </Row>
    );
  }

  return null;
};

export default R2S;
