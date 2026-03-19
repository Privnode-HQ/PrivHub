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

import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../../../helpers';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import {
  Avatar,
  Button,
  Card,
  Col,
  Form,
  Row,
  SideSheet,
  Space,
  Spin,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconClose, IconGift, IconSave } from '@douyinfe/semi-icons';

const { Text, Title } = Typography;

const getInitValues = () => ({
  name: '',
  bound_user_id: undefined,
  deduction_amount: 1,
  valid_from: new Date(),
  expires_at: null,
});

const EditTopupCouponModal = ({
  refresh,
  editingCoupon,
  visible,
  handleClose,
}) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);
  const isEdit = editingCoupon?.id !== undefined;
  const [loading, setLoading] = useState(isEdit);

  const loadCoupon = async () => {
    if (!editingCoupon?.id) return;
    try {
      setLoading(true);
      const res = await API.get(`/api/topup-coupon/${editingCoupon.id}`);
      const { success, message, data } = res.data;
      if (success) {
        formApiRef.current?.setValues({
          ...getInitValues(),
          ...data,
          valid_from: data.valid_from
            ? new Date(data.valid_from * 1000)
            : new Date(),
          expires_at: data.expires_at
            ? new Date(data.expires_at * 1000)
            : null,
        });
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
      loadCoupon();
      return;
    }
    setLoading(false);
    formApiRef.current.reset();
    formApiRef.current.setValues(getInitValues());
  }, [visible, isEdit, editingCoupon?.id]);

  const submit = async (values) => {
    try {
      setLoading(true);
      const payload = {
        name: values.name,
        bound_user_id: parseInt(values.bound_user_id),
        deduction_amount: Number(values.deduction_amount),
        valid_from: values.valid_from
          ? Math.floor(values.valid_from.getTime() / 1000)
          : 0,
        expires_at: values.expires_at
          ? Math.floor(values.expires_at.getTime() / 1000)
          : 0,
      };

      let res;
      if (isEdit) {
        res = await API.put('/api/topup-coupon/', {
          ...payload,
          id: editingCoupon.id,
        });
      } else {
        res = await API.post('/api/topup-coupon/', payload);
      }

      const { success, message } = res.data;
      if (success) {
        showSuccess(isEdit ? t('优惠券更新成功') : t('优惠券发放成功'));
        await refresh();
        handleClose();
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
            {isEdit ? t('更新') : t('新建')}
          </Tag>
          <Title heading={4} className='m-0'>
            {isEdit ? t('更新优惠券') : t('发放充值优惠券')}
          </Title>
        </Space>
      }
      bodyStyle={{ padding: '0' }}
      visible={visible}
      width={isMobile ? '100%' : 600}
      footer={
        <div className='flex justify-end bg-white'>
          <Space>
            <Button
              theme='solid'
              onClick={() => formApiRef.current?.submitForm()}
              icon={<IconSave />}
              loading={loading}
            >
              {t('提交')}
            </Button>
            <Button
              theme='light'
              type='primary'
              onClick={handleClose}
              icon={<IconClose />}
            >
              {t('取消')}
            </Button>
          </Space>
        </div>
      }
      closeIcon={null}
      onCancel={handleClose}
    >
      <Spin spinning={loading}>
        <Form
          initValues={getInitValues()}
          getFormApi={(api) => (formApiRef.current = api)}
          onSubmit={submit}
        >
          <div className='p-2'>
            <Card className='!rounded-2xl shadow-sm border-0 mb-6'>
              <div className='flex items-center mb-2'>
                <Avatar size='small' color='green' className='mr-2 shadow-md'>
                  <IconGift size={16} />
                </Avatar>
                <div>
                  <Text className='text-lg font-medium'>{t('基本信息')}</Text>
                  <div className='text-xs text-gray-600'>
                    {t('配置用户、金额和生效时间')}
                  </div>
                </div>
              </div>

              <Row gutter={12}>
                <Col span={24}>
                  <Form.Input
                    field='name'
                    label={t('名称')}
                    placeholder={t('请输入名称')}
                    rules={[{ required: true, message: t('请输入名称') }]}
                    showClear
                  />
                </Col>
                <Col span={24}>
                  <Form.InputNumber
                    field='bound_user_id'
                    label={t('绑定用户 ID')}
                    placeholder={t('请输入用户 ID')}
                    min={1}
                    precision={0}
                    rules={[
                      { required: true, message: t('请输入用户 ID') },
                    ]}
                  />
                </Col>
                <Col span={24}>
                  <Form.InputNumber
                    field='deduction_amount'
                    label={t('优惠金额')}
                    placeholder={t('请输入优惠金额')}
                    min={0.01}
                    step={0.01}
                    rules={[
                      { required: true, message: t('请输入优惠金额') },
                    ]}
                  />
                </Col>
                <Col span={24}>
                  <Form.DatePicker
                    field='valid_from'
                    label={t('生效时间')}
                    type='dateTime'
                    style={{ width: '100%' }}
                  />
                </Col>
                <Col span={24}>
                  <Form.DatePicker
                    field='expires_at'
                    label={t('过期时间')}
                    type='dateTime'
                    placeholder={t('留空为永不过期')}
                    style={{ width: '100%' }}
                  />
                </Col>
              </Row>
            </Card>
          </div>
        </Form>
      </Spin>
    </SideSheet>
  );
};

export default EditTopupCouponModal;
