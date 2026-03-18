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

import React, { useRef } from 'react';
import { Button, Form } from '@douyinfe/semi-ui';
import { IconSearch } from '@douyinfe/semi-icons';
import { TOPUP_COUPON_STATUS } from '../../../constants/topup-coupon.constants';

const TopupCouponsFilters = ({
  formInitValues,
  setFormApi,
  searchTopupCoupons,
  loading,
  searching,
  t,
}) => {
  const formApiRef = useRef(null);

  const handleReset = () => {
    if (!formApiRef.current) return;
    formApiRef.current.reset();
    setTimeout(() => {
      searchTopupCoupons();
    }, 100);
  };

  return (
    <Form
      initValues={formInitValues}
      getFormApi={(api) => {
        setFormApi(api);
        formApiRef.current = api;
      }}
      onSubmit={searchTopupCoupons}
      allowEmpty={true}
      autoComplete='off'
      trigger='change'
      stopValidateWithError={false}
      className='w-full'
    >
      <div className='flex flex-col lg:flex-row items-center gap-2 w-full'>
        <div className='relative w-full lg:w-64'>
          <Form.Input
            field='searchKeyword'
            prefix={<IconSearch />}
            placeholder={t('关键字(id或名称)')}
            showClear
            pure
            size='small'
          />
        </div>
        <div className='w-full lg:w-44'>
          <Form.Select field='status' placeholder={t('状态')} size='small'>
            <Form.Select.Option value=''>
              {t('全部状态')}
            </Form.Select.Option>
            <Form.Select.Option value={TOPUP_COUPON_STATUS.AVAILABLE}>
              {t('可用')}
            </Form.Select.Option>
            <Form.Select.Option value={TOPUP_COUPON_STATUS.RESERVED}>
              {t('占用中')}
            </Form.Select.Option>
            <Form.Select.Option value={TOPUP_COUPON_STATUS.USED}>
              {t('已使用')}
            </Form.Select.Option>
            <Form.Select.Option value={TOPUP_COUPON_STATUS.EXPIRED}>
              {t('已过期')}
            </Form.Select.Option>
            <Form.Select.Option value={TOPUP_COUPON_STATUS.REVOKED}>
              {t('已撤销')}
            </Form.Select.Option>
          </Form.Select>
        </div>
        <div className='w-full lg:w-40'>
          <Form.InputNumber
            field='userId'
            placeholder={t('用户 ID')}
            size='small'
            min={1}
            precision={0}
          />
        </div>
        <div className='flex gap-2 w-full lg:w-auto'>
          <Button
            type='tertiary'
            htmlType='submit'
            loading={loading || searching}
            className='flex-1 lg:flex-initial'
            size='small'
          >
            {t('查询')}
          </Button>
          <Button
            type='tertiary'
            onClick={handleReset}
            className='flex-1 lg:flex-initial'
            size='small'
          >
            {t('重置')}
          </Button>
        </div>
      </div>
    </Form>
  );
};

export default TopupCouponsFilters;
