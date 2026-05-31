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
import {
  Avatar,
  Button,
  Col,
  Form,
  Input,
  Modal,
  Row,
  Select,
  SideSheet,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IconClose,
  IconPlus,
  IconSave,
  IconSearch,
} from '@douyinfe/semi-icons';
import { BadgePercent, TicketPercent } from 'lucide-react';
import CardPro from '../../common/ui/CardPro';
import CardTable from '../../common/ui/CardTable';
import {
  API,
  formatCurrencyAmountByCode,
  showError,
  showSuccess,
  timestamp2string,
} from '../../../helpers';
import { createCardProPagination } from '../../../helpers/utils';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { ITEMS_PER_PAGE } from '../../../constants';
import {
  TOPUP_PROMOTION_REDEMPTION_STATUS_MAP,
  TOPUP_PROMOTION_STATUS_MAP,
} from '../../../constants/topup-coupon.constants';

const { Text, Title } = Typography;

const defaultRule = () => ({
  min_amount: 0,
  discount_type: 'fixed',
  discount_value: 1,
});

const getDefaultCurrencyCode = () => {
  const quotaDisplayType = (localStorage.getItem('quota_display_type') || 'USD')
    .trim()
    .toUpperCase();
  return quotaDisplayType === 'CNY' ? 'CNY' : 'USD';
};

const toTimestamp = (value) => {
  if (!value) return 0;
  return Math.floor(value.getTime() / 1000);
};

const fromTimestamp = (value) => {
  if (!value) return null;
  return new Date(value * 1000);
};

const splitCodes = (value) => {
  return String(value || '')
    .split(/[\s,，]+/)
    .map((code) => code.trim().toUpperCase())
    .filter(Boolean);
};

const renderStatus = (status, map, t) => {
  const config = map[status] || { color: 'grey', text: '未知状态' };
  return (
    <Tag color={config.color} shape='circle'>
      {t(config.text)}
    </Tag>
  );
};

const renderTime = (timestamp, emptyText, t) => {
  if (!timestamp) return t(emptyText);
  return timestamp2string(timestamp);
};

const formatRule = (rule, currencyCode, t) => {
  if (!rule) return '';
  const amount = formatCurrencyAmountByCode(rule.min_amount || 0, currencyCode);
  if (rule.discount_type === 'percent') {
    return `${t('满')} ${amount} ${t('减')} ${rule.discount_value}%`;
  }
  return `${t('满')} ${amount} ${t('减')} ${formatCurrencyAmountByCode(
    rule.discount_value || 0,
    currencyCode,
  )}`;
};

const PromotionEditorSheet = ({ visible, campaign, onClose, onSaved, t }) => {
  const formApiRef = useRef(null);
  const isEdit = Boolean(campaign?.id);
  const [loading, setLoading] = useState(false);
  const [rules, setRules] = useState([defaultRule()]);

  const setInitialValues = (data = {}) => {
    setRules(data.rules?.length ? data.rules : [defaultRule()]);
    formApiRef.current?.setValues({
      name: data.name || '',
      description: data.description || '',
      currency_code: data.currency_code || getDefaultCurrencyCode(),
      allowed_groups: data.allowed_groups || [],
      valid_from: data.valid_from ? fromTimestamp(data.valid_from) : new Date(),
      expires_at: data.expires_at ? fromTimestamp(data.expires_at) : null,
      max_redemptions: data.max_redemptions || 0,
      max_redemptions_per_user: data.max_redemptions_per_user || 0,
      codes_text: '',
      auto_code_count: isEdit ? 0 : 1,
      code_prefix: '',
      code_valid_from: null,
      code_expires_at: null,
      code_max_redemptions: 0,
      code_max_redemptions_per_user: 0,
    });
  };

  const loadCampaign = async () => {
    if (!campaign?.id) return;
    setLoading(true);
    try {
      const res = await API.get(`/api/topup-promotion/${campaign.id}`);
      const { success, message, data } = res.data;
      if (success) {
        setInitialValues(data);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!visible || !formApiRef.current) return;
    if (isEdit) {
      loadCampaign();
      return;
    }
    setInitialValues();
  }, [visible, isEdit, campaign?.id]);

  const updateRule = (index, patch) => {
    setRules((current) =>
      current.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    );
  };

  const submit = async (values) => {
    const payload = {
      name: values.name,
      description: values.description,
      currency_code: String(values.currency_code || '')
        .trim()
        .toUpperCase(),
      rules: rules.map((rule) => ({
        min_amount: Number(rule.min_amount) || 0,
        discount_type: rule.discount_type || 'fixed',
        discount_value: Number(rule.discount_value) || 0,
      })),
      allowed_groups: values.allowed_groups || [],
      valid_from: toTimestamp(values.valid_from),
      expires_at: toTimestamp(values.expires_at),
      max_redemptions: Number(values.max_redemptions) || 0,
      max_redemptions_per_user: Number(values.max_redemptions_per_user) || 0,
    };
    if (!isEdit) {
      payload.codes = splitCodes(values.codes_text);
      payload.auto_code_count = Number(values.auto_code_count) || 0;
      payload.code_prefix = values.code_prefix || '';
      payload.code_valid_from = toTimestamp(values.code_valid_from);
      payload.code_expires_at = toTimestamp(values.code_expires_at);
      payload.code_max_redemptions = Number(values.code_max_redemptions) || 0;
      payload.code_max_redemptions_per_user =
        Number(values.code_max_redemptions_per_user) || 0;
    }

    setLoading(true);
    try {
      const res = isEdit
        ? await API.put('/api/topup-promotion/', {
            ...payload,
            id: campaign.id,
          })
        : await API.post('/api/topup-promotion/', payload);
      const { success, message } = res.data;
      if (success) {
        showSuccess(isEdit ? t('促销活动已更新') : t('促销活动已创建'));
        onSaved();
        onClose();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <SideSheet
      placement={isEdit ? 'right' : 'left'}
      title={
        <Space>
          <Tag color={isEdit ? 'blue' : 'green'} shape='circle'>
            {isEdit ? t('编辑') : t('新建')}
          </Tag>
          <Title heading={4} className='m-0'>
            {isEdit ? t('编辑促销活动') : t('创建促销活动')}
          </Title>
        </Space>
      }
      visible={visible}
      width={720}
      bodyStyle={{ padding: 0 }}
      closeIcon={null}
      onCancel={onClose}
      footer={
        <div className='flex justify-end bg-white'>
          <Space>
            <Button
              theme='solid'
              icon={<IconSave />}
              loading={loading}
              onClick={() => formApiRef.current?.submitForm()}
            >
              {t('保存')}
            </Button>
            <Button theme='light' icon={<IconClose />} onClick={onClose}>
              {t('取消')}
            </Button>
          </Space>
        </div>
      }
    >
      <Spin spinning={loading}>
        <Form
          getFormApi={(api) => (formApiRef.current = api)}
          onSubmit={submit}
          initValues={{ currency_code: getDefaultCurrencyCode() }}
        >
          <div className='p-4 space-y-5'>
            <section>
              <div className='mb-3'>
                <Text strong>{t('活动信息')}</Text>
                <div className='text-xs text-gray-500'>
                  {t('促销码只会在易支付结账时生效')}
                </div>
              </div>
              <Row gutter={12}>
                <Col span={16}>
                  <Form.Input
                    field='name'
                    label={t('活动名称')}
                    rules={[{ required: true, message: t('请输入活动名称') }]}
                    showClear
                  />
                </Col>
                <Col span={8}>
                  <Form.Input
                    field='currency_code'
                    label={t('结算货币')}
                    maxLength={16}
                    rules={[{ required: true, message: t('请输入结算货币') }]}
                  />
                </Col>
                <Col span={24}>
                  <Form.TextArea
                    field='description'
                    label={t('说明')}
                    rows={2}
                    placeholder={t('给运营人员看的活动说明')}
                    showClear
                  />
                </Col>
                <Col span={24}>
                  <Form.TagInput
                    field='allowed_groups'
                    label={t('可兑换用户分组')}
                    placeholder={t('留空代表全部分组，输入后回车')}
                    addOnBlur
                    showClear
                  />
                </Col>
                <Col span={12}>
                  <Form.DatePicker
                    field='valid_from'
                    label={t('活动生效时间')}
                    type='dateTime'
                    style={{ width: '100%' }}
                  />
                </Col>
                <Col span={12}>
                  <Form.DatePicker
                    field='expires_at'
                    label={t('活动过期时间')}
                    type='dateTime'
                    placeholder={t('留空为永不过期')}
                    style={{ width: '100%' }}
                  />
                </Col>
                <Col span={12}>
                  <Form.InputNumber
                    field='max_redemptions'
                    label={t('活动总兑换次数')}
                    min={0}
                    precision={0}
                    extraText={t('0 表示不限制')}
                  />
                </Col>
                <Col span={12}>
                  <Form.InputNumber
                    field='max_redemptions_per_user'
                    label={t('每用户活动兑换次数')}
                    min={0}
                    precision={0}
                    extraText={t('0 表示不限制')}
                  />
                </Col>
              </Row>
            </section>

            <section>
              <div className='flex justify-between items-center mb-3'>
                <div>
                  <Text strong>{t('阶梯促销规则')}</Text>
                  <div className='text-xs text-gray-500'>
                    {t('系统会使用满足门槛的最高一档规则')}
                  </div>
                </div>
                <Button
                  icon={<IconPlus />}
                  type='tertiary'
                  size='small'
                  onClick={() =>
                    setRules((current) => [...current, defaultRule()])
                  }
                >
                  {t('添加阶梯')}
                </Button>
              </div>
              <div className='space-y-2'>
                {rules.map((rule, index) => (
                  <div
                    key={index}
                    className='grid grid-cols-12 gap-2 items-end rounded-md border p-2'
                    style={{ borderColor: 'var(--semi-color-border)' }}
                  >
                    <Input
                      className='col-span-3'
                      prefix={t('满')}
                      value={rule.min_amount}
                      onChange={(value) =>
                        updateRule(index, { min_amount: Number(value) || 0 })
                      }
                    />
                    <Select
                      className='col-span-3'
                      value={rule.discount_type}
                      onChange={(value) =>
                        updateRule(index, { discount_type: value })
                      }
                    >
                      <Select.Option value='fixed'>
                        {t('固定金额')}
                      </Select.Option>
                      <Select.Option value='percent'>
                        {t('百分比')}
                      </Select.Option>
                    </Select>
                    <Input
                      className='col-span-4'
                      prefix={t('减')}
                      suffix={rule.discount_type === 'percent' ? '%' : null}
                      value={rule.discount_value}
                      onChange={(value) =>
                        updateRule(index, {
                          discount_value: Number(value) || 0,
                        })
                      }
                    />
                    <Button
                      className='col-span-2'
                      type='danger'
                      theme='borderless'
                      disabled={rules.length === 1}
                      onClick={() =>
                        setRules((current) =>
                          current.filter((_, ruleIndex) => ruleIndex !== index),
                        )
                      }
                    >
                      {t('移除')}
                    </Button>
                  </div>
                ))}
              </div>
            </section>

            {!isEdit && (
              <section>
                <div className='mb-3'>
                  <Text strong>{t('初始促销码')}</Text>
                  <div className='text-xs text-gray-500'>
                    {t('可输入自定义促销码，也可自动生成多个促销码')}
                  </div>
                </div>
                <Row gutter={12}>
                  <Col span={24}>
                    <Form.TextArea
                      field='codes_text'
                      label={t('自定义促销码')}
                      rows={2}
                      placeholder={t('多个促销码用空格、逗号或换行分隔')}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.InputNumber
                      field='auto_code_count'
                      label={t('自动生成数量')}
                      min={0}
                      max={500}
                      precision={0}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.Input field='code_prefix' label={t('促销码前缀')} />
                  </Col>
                  <Col span={12}>
                    <Form.DatePicker
                      field='code_valid_from'
                      label={t('促销码生效时间')}
                      type='dateTime'
                      placeholder={t('默认使用活动生效时间')}
                      style={{ width: '100%' }}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.DatePicker
                      field='code_expires_at'
                      label={t('促销码过期时间')}
                      type='dateTime'
                      placeholder={t('默认使用活动过期时间')}
                      style={{ width: '100%' }}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.InputNumber
                      field='code_max_redemptions'
                      label={t('单个促销码总次数')}
                      min={0}
                      precision={0}
                      extraText={t('0 表示不限制')}
                    />
                  </Col>
                  <Col span={12}>
                    <Form.InputNumber
                      field='code_max_redemptions_per_user'
                      label={t('每用户促销码次数')}
                      min={0}
                      precision={0}
                      extraText={t('0 表示不限制')}
                    />
                  </Col>
                </Row>
              </section>
            )}
          </div>
        </Form>
      </Spin>
    </SideSheet>
  );
};

const PromotionCodesModal = ({ campaign, visible, onCancel, t }) => {
  const formApiRef = useRef(null);
  const [codes, setCodes] = useState([]);
  const [loading, setLoading] = useState(false);

  const loadCodes = async () => {
    if (!campaign?.id) return;
    setLoading(true);
    try {
      const res = await API.get(
        `/api/topup-promotion/codes?campaign_id=${campaign.id}&page_size=100`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setCodes(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadCodes();
    }
  }, [visible, campaign?.id]);

  const addCodes = async (values) => {
    setLoading(true);
    try {
      const res = await API.post('/api/topup-promotion/codes', {
        campaign_id: campaign.id,
        codes: splitCodes(values.codes_text),
        auto_code_count: Number(values.auto_code_count) || 0,
        code_prefix: values.code_prefix || '',
        valid_from: toTimestamp(values.valid_from),
        expires_at: toTimestamp(values.expires_at),
        max_redemptions: Number(values.max_redemptions) || 0,
        max_redemptions_per_user: Number(values.max_redemptions_per_user) || 0,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('促销码已添加'));
        formApiRef.current?.reset();
        await loadCodes();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const revokeCode = async (code) => {
    Modal.confirm({
      title: t('撤销促销码'),
      content: `${t('撤销后该促销码不能再用于新订单')}：${code.code}`,
      okButtonProps: { type: 'danger' },
      okText: t('撤销促销码'),
      cancelText: t('取消'),
      onOk: async () => {
        const res = await API.put('/api/topup-promotion/codes', {
          id: code.id,
          action: 'revoke',
        });
        const { success, message } = res.data;
        if (success) {
          showSuccess(t('促销码已撤销'));
          loadCodes();
        } else {
          showError(message);
        }
      },
    });
  };

  const columns = [
    { title: t('促销码'), dataIndex: 'code' },
    {
      title: t('状态'),
      dataIndex: 'effective_status',
      render: (text, record) =>
        renderStatus(text || record.status, TOPUP_PROMOTION_STATUS_MAP, t),
    },
    {
      title: t('使用'),
      dataIndex: 'used_count',
      render: (text, record) =>
        `${record.used_count || 0}/${record.max_redemptions || t('不限')}`,
    },
    {
      title: t('有效期'),
      dataIndex: 'valid_from',
      render: (text, record) =>
        `${renderTime(text, '立即', t)} - ${renderTime(
          record.expires_at,
          '永不过期',
          t,
        )}`,
    },
    {
      title: '',
      dataIndex: 'operate',
      render: (_, record) => (
        <Button
          type='danger'
          size='small'
          disabled={(record.effective_status || record.status) === 'revoked'}
          onClick={() => revokeCode(record)}
        >
          {t('撤销')}
        </Button>
      ),
    },
  ];

  return (
    <Modal
      title={`${t('促销码')} - ${campaign?.name || ''}`}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={900}
    >
      <Form
        getFormApi={(api) => (formApiRef.current = api)}
        onSubmit={addCodes}
        initValues={{ auto_code_count: 1 }}
      >
        <div className='grid grid-cols-1 md:grid-cols-6 gap-2 mb-3'>
          <Form.TextArea
            field='codes_text'
            placeholder={t('自定义促销码')}
            rows={1}
            noLabel
            className='md:col-span-2'
          />
          <Form.InputNumber
            field='auto_code_count'
            placeholder={t('生成数量')}
            min={0}
            max={500}
            precision={0}
            noLabel
          />
          <Form.Input field='code_prefix' placeholder={t('前缀')} noLabel />
          <Form.InputNumber
            field='max_redemptions'
            placeholder={t('总次数')}
            min={0}
            precision={0}
            noLabel
          />
          <Button htmlType='submit' icon={<IconPlus />} loading={loading}>
            {t('添加')}
          </Button>
        </div>
      </Form>
      <Table
        rowKey='id'
        columns={columns}
        dataSource={codes}
        loading={loading}
        pagination={false}
      />
    </Modal>
  );
};

const PromotionRedemptionsModal = ({ campaign, visible, onCancel, t }) => {
  const [redemptions, setRedemptions] = useState([]);
  const [loading, setLoading] = useState(false);

  const loadRedemptions = async () => {
    if (!campaign?.id) return;
    setLoading(true);
    try {
      const res = await API.get(
        `/api/topup-promotion/redemptions?campaign_id=${campaign.id}&page_size=100`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setRedemptions(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadRedemptions();
    }
  }, [visible, campaign?.id]);

  const columns = [
    { title: t('促销码'), dataIndex: 'code' },
    {
      title: t('用户'),
      dataIndex: 'username',
      render: (text, record) => text || `#${record.user_id}`,
    },
    { title: t('订单号'), dataIndex: 'trade_no' },
    {
      title: t('抵扣'),
      dataIndex: 'discount_amount',
      render: (text, record) =>
        formatCurrencyAmountByCode(text, record.currency_code),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (status) =>
        renderStatus(status, TOPUP_PROMOTION_REDEMPTION_STATUS_MAP, t),
    },
    {
      title: t('核销时间'),
      dataIndex: 'used_at',
      render: (text, record) => renderTime(text || record.reserved_at, '无', t),
    },
  ];

  return (
    <Modal
      title={`${t('使用记录')} - ${campaign?.name || ''}`}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={980}
    >
      <Table
        rowKey='id'
        columns={columns}
        dataSource={redemptions}
        loading={loading}
        pagination={false}
      />
    </Modal>
  );
};

const TopupPromotionsPanel = ({ readOnlyAdmin, t }) => {
  const isMobile = useIsMobile();
  const [campaigns, setCampaigns] = useState([]);
  const [loading, setLoading] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [total, setTotal] = useState(0);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [editingCampaign, setEditingCampaign] = useState(null);
  const [showEditor, setShowEditor] = useState(false);
  const [codesCampaign, setCodesCampaign] = useState(null);
  const [recordsCampaign, setRecordsCampaign] = useState(null);

  const loadCampaigns = async (page = activePage, size = pageSize) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        p: String(page),
        page_size: String(size),
      });
      if (keyword) params.set('keyword', keyword);
      if (status) params.set('status', status);
      const res = await API.get(`/api/topup-promotion/?${params.toString()}`);
      const { success, message, data } = res.data;
      if (success) {
        setCampaigns(data.items || []);
        setActivePage(data.page || 1);
        setTotal(data.total || 0);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadCampaigns(1, pageSize);
  }, [pageSize]);

  const revokeCampaign = (campaign) => {
    Modal.confirm({
      title: t('撤销促销活动'),
      content: `${t('撤销后该活动下的促销码不能再用于新订单')}：${campaign.name}`,
      okButtonProps: { type: 'danger' },
      okText: t('撤销促销活动'),
      cancelText: t('取消'),
      onOk: async () => {
        const res = await API.put('/api/topup-promotion/', {
          id: campaign.id,
          action: 'revoke',
        });
        const { success, message } = res.data;
        if (success) {
          showSuccess(t('促销活动已撤销'));
          loadCampaigns();
        } else {
          showError(message);
        }
      },
    });
  };

  const columns = useMemo(
    () => [
      { title: t('ID'), dataIndex: 'id' },
      { title: t('活动名称'), dataIndex: 'name' },
      {
        title: t('状态'),
        dataIndex: 'effective_status',
        render: (text, record) =>
          renderStatus(text || record.status, TOPUP_PROMOTION_STATUS_MAP, t),
      },
      {
        title: t('促销规则'),
        dataIndex: 'rules',
        render: (rules, record) => (
          <div className='flex flex-col gap-1'>
            {(rules || []).slice(0, 2).map((rule, index) => (
              <Text key={index} size='small'>
                {formatRule(rule, record.currency_code, t)}
              </Text>
            ))}
            {(rules || []).length > 2 && (
              <Text type='tertiary' size='small'>
                {t('还有')} {(rules || []).length - 2} {t('档')}
              </Text>
            )}
          </div>
        ),
      },
      {
        title: t('适用分组'),
        dataIndex: 'allowed_groups',
        render: (groups) =>
          groups?.length ? groups.join(', ') : t('全部分组'),
      },
      {
        title: t('促销码'),
        dataIndex: 'code_count',
        render: (text, record) =>
          `${text || 0} (${t('已核销')} ${record.used_count || 0})`,
      },
      {
        title: t('有效期'),
        dataIndex: 'valid_from',
        render: (text, record) =>
          `${renderTime(text, '立即', t)} - ${renderTime(
            record.expires_at,
            '永不过期',
            t,
          )}`,
      },
      {
        title: '',
        dataIndex: 'operate',
        fixed: 'right',
        render: (_, record) => (
          <Space>
            <Button size='small' onClick={() => setCodesCampaign(record)}>
              {t('促销码')}
            </Button>
            <Button size='small' onClick={() => setRecordsCampaign(record)}>
              {t('记录')}
            </Button>
            {!readOnlyAdmin && (
              <>
                <Button
                  size='small'
                  type='tertiary'
                  disabled={
                    (record.effective_status || record.status) === 'revoked'
                  }
                  onClick={() => {
                    setEditingCampaign(record);
                    setShowEditor(true);
                  }}
                >
                  {t('编辑')}
                </Button>
                <Button
                  size='small'
                  type='danger'
                  disabled={
                    (record.effective_status || record.status) === 'revoked'
                  }
                  onClick={() => revokeCampaign(record)}
                >
                  {t('撤销')}
                </Button>
              </>
            )}
          </Space>
        ),
      },
    ],
    [t, readOnlyAdmin],
  );

  return (
    <>
      <PromotionEditorSheet
        visible={showEditor}
        campaign={editingCampaign}
        onClose={() => {
          setShowEditor(false);
          setEditingCampaign(null);
        }}
        onSaved={() => loadCampaigns()}
        t={t}
      />
      <PromotionCodesModal
        campaign={codesCampaign}
        visible={Boolean(codesCampaign)}
        onCancel={() => setCodesCampaign(null)}
        t={t}
      />
      <PromotionRedemptionsModal
        campaign={recordsCampaign}
        visible={Boolean(recordsCampaign)}
        onCancel={() => setRecordsCampaign(null)}
        t={t}
      />
      <CardPro
        type='type1'
        descriptionArea={
          <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
            <div className='flex items-center text-emerald-500'>
              <Avatar size='small' color='green' className='mr-2 shadow-md'>
                <BadgePercent size={14} />
              </Avatar>
              <Text>{t('促销活动')}</Text>
            </div>
          </div>
        }
        actionsArea={
          <div className='flex flex-col lg:flex-row justify-between items-center gap-2 w-full'>
            <div className='flex gap-2 w-full lg:w-auto'>
              {!readOnlyAdmin && (
                <Button
                  type='primary'
                  icon={<TicketPercent size={14} />}
                  onClick={() => {
                    setEditingCampaign(null);
                    setShowEditor(true);
                  }}
                >
                  {t('创建促销活动')}
                </Button>
              )}
              <Button type='tertiary' onClick={() => loadCampaigns()}>
                {t('刷新')}
              </Button>
            </div>
            <div className='flex flex-col md:flex-row gap-2 w-full lg:w-auto'>
              <Input
                prefix={<IconSearch />}
                value={keyword}
                onChange={setKeyword}
                placeholder={t('搜索活动名称或 ID')}
                showClear
              />
              <Select
                value={status}
                onChange={setStatus}
                placeholder={t('状态')}
                style={{ minWidth: 140 }}
              >
                <Select.Option value=''>{t('全部状态')}</Select.Option>
                <Select.Option value='active'>{t('生效中')}</Select.Option>
                <Select.Option value='paused'>{t('已暂停')}</Select.Option>
                <Select.Option value='revoked'>{t('已撤销')}</Select.Option>
              </Select>
              <Button onClick={() => loadCampaigns(1, pageSize)}>
                {t('查询')}
              </Button>
            </div>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: activePage,
          pageSize,
          total,
          onPageChange: (page) => {
            setActivePage(page);
            loadCampaigns(page, pageSize);
          },
          onPageSizeChange: (size) => {
            setPageSize(size);
            setActivePage(1);
            loadCampaigns(1, size);
          },
          isMobile,
          t,
        })}
        t={t}
      >
        <CardTable
          rowKey='id'
          columns={columns}
          dataSource={campaigns}
          loading={loading}
          hidePagination={true}
          scroll={{ x: 'max-content' }}
        />
      </CardPro>
    </>
  );
};

export default TopupPromotionsPanel;
