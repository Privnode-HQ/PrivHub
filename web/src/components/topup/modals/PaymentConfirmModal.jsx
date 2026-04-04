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

import React from 'react';
import { Modal, Typography, Card, Select, Skeleton } from '@douyinfe/semi-ui';
import { SiAlipay, SiWechat, SiStripe } from 'react-icons/si';
import { CreditCard } from 'lucide-react';
import { formatCurrencyAmountByCode } from '../../../helpers';

const { Text } = Typography;

const PaymentConfirmModal = ({
  t,
  open,
  onlineTopUp,
  handleCancel,
  confirmLoading,
  topUpCount,
  renderQuotaWithAmount,
  amountLoading,
  payWay,
  payMethods,
  currencyCode,
  supportedCurrencyCodes,
  selectedCurrencyCode,
  onCurrencyChange,
  originalAmount,
  platformDiscountAmount,
  couponDiscountAmount,
  finalPayableAmount,
  availableCoupons,
  selectedCouponId,
  onCouponChange,
  ineligibleReason,
}) => {
  const normalizedOriginalAmount =
    originalAmount && originalAmount > 0 ? originalAmount : 0;
  const normalizedPlatformDiscount =
    platformDiscountAmount && platformDiscountAmount > 0
      ? platformDiscountAmount
      : 0;
  const normalizedCouponDiscount =
    couponDiscountAmount && couponDiscountAmount > 0 ? couponDiscountAmount : 0;
  const normalizedFinalAmount =
    finalPayableAmount && finalPayableAmount > 0 ? finalPayableAmount : 0;
  const hasAnyDiscount =
    normalizedPlatformDiscount > 0 || normalizedCouponDiscount > 0;
  const showStripeCurrencySelector =
    payWay === 'stripe' && (supportedCurrencyCodes?.length || 0) > 1;
  return (
    <Modal
      title={
        <div className='flex items-center'>
          <CreditCard className='mr-2' size={18} />
          {t('充值确认')}
        </div>
      }
      visible={open}
      onOk={onlineTopUp}
      onCancel={handleCancel}
      maskClosable={false}
      size='small'
      centered
      confirmLoading={confirmLoading}
      okButtonProps={{
        disabled: amountLoading || Boolean(ineligibleReason),
      }}
    >
      <div className='space-y-4'>
        <Card className='!rounded-xl !border-0 bg-slate-50 dark:bg-slate-800'>
          <div className='space-y-3'>
            <div className='flex justify-between items-center'>
              <Text strong className='text-slate-700 dark:text-slate-200'>
                {t('充值数量')}：
              </Text>
              <Text className='text-slate-900 dark:text-slate-100'>
                {renderQuotaWithAmount(topUpCount)}
              </Text>
            </div>
            <div className='flex justify-between items-center'>
              <Text strong className='text-slate-700 dark:text-slate-200'>
                {t('实付金额')}：
              </Text>
              {amountLoading ? (
                <Skeleton.Title style={{ width: '60px', height: '16px' }} />
              ) : (
                <Text strong className='font-bold' style={{ color: 'red' }}>
                  {formatCurrencyAmountByCode(
                    normalizedFinalAmount,
                    currencyCode,
                  )}
                </Text>
              )}
            </div>
            {showStripeCurrencySelector && (
              <div className='space-y-2'>
                <Text strong className='text-slate-700 dark:text-slate-200'>
                  {t('结算货币')}：
                </Text>
                <Select
                  value={selectedCurrencyCode || currencyCode}
                  onChange={onCurrencyChange}
                  size='small'
                  style={{ width: '100%' }}
                  disabled={amountLoading || confirmLoading}
                >
                  {(supportedCurrencyCodes || []).map(
                    (supportedCurrencyCode) => (
                      <Select.Option
                        value={supportedCurrencyCode}
                        key={supportedCurrencyCode}
                      >
                        {supportedCurrencyCode}
                      </Select.Option>
                    ),
                  )}
                </Select>
              </div>
            )}
            {availableCoupons?.length > 0 && (
              <div className='space-y-2'>
                <Text strong className='text-slate-700 dark:text-slate-200'>
                  {t('选择优惠券')}：
                </Text>
                <Select
                  value={selectedCouponId || 0}
                  onChange={onCouponChange}
                  size='small'
                  style={{ width: '100%' }}
                >
                  <Select.Option value={0}>{t('不使用优惠券')}</Select.Option>
                  {availableCoupons.map((coupon) => (
                    <Select.Option value={coupon.id} key={coupon.id}>
                      {`${coupon.name} (-${formatCurrencyAmountByCode(
                        coupon.deduction_amount,
                        coupon.currency_code || currencyCode,
                      )})`}
                    </Select.Option>
                  ))}
                </Select>
              </div>
            )}
            {ineligibleReason && (
              <Text type='danger'>{t(ineligibleReason)}</Text>
            )}
            {hasAnyDiscount && !amountLoading && (
              <>
                <div className='flex justify-between items-center'>
                  <Text className='text-slate-500 dark:text-slate-400'>
                    {t('原价')}：
                  </Text>
                  <Text delete className='text-slate-500 dark:text-slate-400'>
                    {formatCurrencyAmountByCode(
                      normalizedOriginalAmount,
                      currencyCode,
                    )}
                  </Text>
                </div>
                {normalizedPlatformDiscount > 0 && (
                  <div className='flex justify-between items-center'>
                    <Text className='text-slate-500 dark:text-slate-400'>
                      {t('平台优惠')}：
                    </Text>
                    <Text className='text-emerald-600 dark:text-emerald-400'>
                      -{' '}
                      {formatCurrencyAmountByCode(
                        normalizedPlatformDiscount,
                        currencyCode,
                      )}
                    </Text>
                  </div>
                )}
                {normalizedCouponDiscount > 0 && (
                  <div className='flex justify-between items-center'>
                    <Text className='text-slate-500 dark:text-slate-400'>
                      {t('优惠券抵扣')}：
                    </Text>
                    <Text className='text-emerald-600 dark:text-emerald-400'>
                      -{' '}
                      {formatCurrencyAmountByCode(
                        normalizedCouponDiscount,
                        currencyCode,
                      )}
                    </Text>
                  </div>
                )}
              </>
            )}
            <div className='flex justify-between items-center'>
              <Text strong className='text-slate-700 dark:text-slate-200'>
                {t('支付方式')}：
              </Text>
              <div className='flex items-center'>
                {(() => {
                  const payMethod = payMethods.find(
                    (method) => method.type === payWay,
                  );
                  if (payMethod) {
                    return (
                      <>
                        {payMethod.type === 'alipay' ? (
                          <SiAlipay
                            className='mr-2'
                            size={16}
                            color='#1677FF'
                          />
                        ) : payMethod.type === 'wxpay' ? (
                          <SiWechat
                            className='mr-2'
                            size={16}
                            color='#07C160'
                          />
                        ) : payMethod.type === 'stripe' ? (
                          <SiStripe
                            className='mr-2'
                            size={16}
                            color='#635BFF'
                          />
                        ) : (
                          <CreditCard
                            className='mr-2'
                            size={16}
                            color={
                              payMethod.color || 'var(--semi-color-text-2)'
                            }
                          />
                        )}
                        <Text className='text-slate-900 dark:text-slate-100'>
                          {payMethod.name}
                        </Text>
                      </>
                    );
                  } else {
                    // 默认充值方式
                    if (payWay === 'alipay') {
                      return (
                        <>
                          <SiAlipay
                            className='mr-2'
                            size={16}
                            color='#1677FF'
                          />
                          <Text className='text-slate-900 dark:text-slate-100'>
                            {t('支付宝')}
                          </Text>
                        </>
                      );
                    } else if (payWay === 'stripe') {
                      return (
                        <>
                          <SiStripe
                            className='mr-2'
                            size={16}
                            color='#635BFF'
                          />
                          <Text className='text-slate-900 dark:text-slate-100'>
                            Stripe
                          </Text>
                        </>
                      );
                    } else {
                      return (
                        <>
                          <SiWechat
                            className='mr-2'
                            size={16}
                            color='#07C160'
                          />
                          <Text className='text-slate-900 dark:text-slate-100'>
                            {t('微信')}
                          </Text>
                        </>
                      );
                    }
                  }
                })()}
              </div>
            </div>
          </div>
        </Card>
      </div>
    </Modal>
  );
};

export default PaymentConfirmModal;
