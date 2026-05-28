import React, { useMemo } from 'react';
import { Form, InputNumber, Typography, withField } from '@douyinfe/semi-ui';
import {
  currencyAmountToQuota,
  getCurrencyConfig,
  getQuotaPerUnit,
  quotaToCurrencyAmount,
} from '../../helpers';

const { Text } = Typography;

function amountFromQuota(quota, precision) {
  if (quota === '' || quota === undefined || quota === null) {
    return undefined;
  }
  const amount = quotaToCurrencyAmount(quota);
  if (!Number.isFinite(amount)) {
    return undefined;
  }
  return Number(amount.toFixed(precision));
}

function resolveAmountPrecision(precision) {
  if (Number.isInteger(precision) && precision >= 0) {
    return precision;
  }
  const quotaPerUnit = getQuotaPerUnit();
  if (!Number.isFinite(quotaPerUnit) || quotaPerUnit <= 1) {
    return 6;
  }
  return Math.max(6, Math.ceil(Math.log10(quotaPerUnit)));
}

function QuotaAmountControl({
  value,
  onChange,
  disabled = false,
  minQuota,
  maxQuota,
  allowNegative = false,
  precision,
  step = 1,
  placeholder,
  size,
  style,
  className,
  showClear = true,
  extraText,
  ...inputProps
}) {
  const { symbol, type } = getCurrencyConfig();
  const amountPrecision = resolveAmountPrecision(precision);
  const displayValue = useMemo(
    () => amountFromQuota(value, amountPrecision),
    [amountPrecision, value],
  );
  const minAmount = allowNegative
    ? undefined
    : minQuota === undefined
      ? 0
      : amountFromQuota(minQuota, amountPrecision);
  const maxAmount =
    maxQuota === undefined
      ? undefined
      : amountFromQuota(maxQuota, amountPrecision);

  return (
    <>
      <InputNumber
        {...inputProps}
        value={displayValue}
        onChange={(amount) => onChange?.(currencyAmountToQuota(amount))}
        disabled={disabled}
        min={minAmount}
        max={maxAmount}
        precision={amountPrecision}
        step={step}
        prefix={symbol}
        suffix={type}
        placeholder={placeholder}
        size={size}
        style={{ width: '100%', ...style }}
        className={className}
        showClear={showClear}
      />
      {extraText ? (
        <Text type='tertiary' size='small' style={{ display: 'block' }}>
          {extraText}
        </Text>
      ) : null}
    </>
  );
}

const FormQuotaAmountControl = withField(QuotaAmountControl);

function QuotaAmountInput({ label, field, rules, ...props }) {
  if (field) {
    return (
      <FormQuotaAmountControl
        field={field}
        label={label}
        rules={rules}
        {...props}
      />
    );
  }

  const input = <QuotaAmountControl {...props} />;

  if (!label) {
    return input;
  }

  return <Form.Slot label={label}>{input}</Form.Slot>;
}

export default QuotaAmountInput;
