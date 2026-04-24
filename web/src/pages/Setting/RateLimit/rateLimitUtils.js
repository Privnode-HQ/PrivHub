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

let multiplierRuleKeySeed = 0;

const nextMultiplierRuleKey = () => {
  multiplierRuleKeySeed += 1;
  return `usage-limit-rule-${Date.now()}-${multiplierRuleKeySeed}`;
};

export const usageLimitTargetOptions = [
  { label: '全部（非封禁）用户', value: 'all' },
  { label: '特定分组', value: 'groups' },
  { label: '特定用户', value: 'users' },
];

export const usageLimitMetricOptions = [
  { label: 'RPM', value: 'rpm' },
  { label: 'RPD', value: 'rpd' },
  { label: 'TPM', value: 'tpm' },
  { label: 'TPD', value: 'tpd' },
  { label: 'Hourly', value: 'hourly' },
  { label: 'Daily', value: 'daily' },
  { label: 'Weekly', value: 'weekly' },
  { label: 'Monthly', value: 'monthly' },
];

const normalizeStringList = (values = []) => {
  const result = [];
  const seen = new Set();

  values.forEach((value) => {
    const normalized = `${value || ''}`.trim();
    if (!normalized || seen.has(normalized)) {
      return;
    }
    seen.add(normalized);
    result.push(normalized);
  });

  return result;
};

const normalizeNumberList = (values = []) => {
  const result = [];
  const seen = new Set();

  values.forEach((value) => {
    const normalized = Number(value);
    if (!Number.isInteger(normalized) || normalized <= 0 || seen.has(normalized)) {
      return;
    }
    seen.add(normalized);
    result.push(normalized);
  });

  return result;
};

export const formatUserOption = (user) => {
  const userLabel = user.display_name || user.username || `#${user.id}`;
  const suffix = [
    user.cah_id ? `(${user.cah_id})` : `(#${user.id})`,
    user.group || '',
    user.email || '',
  ]
    .filter(Boolean)
    .join(' / ');

  return {
    label: suffix ? `${userLabel} ${suffix}` : userLabel,
    value: user.id,
  };
};

export const mergeUserOptions = (currentOptions = [], users = []) => {
  const merged = [...currentOptions];
  const existingIds = new Set(currentOptions.map((item) => item.value));

  users.forEach((user) => {
    const option =
      user && Object.prototype.hasOwnProperty.call(user, 'label')
        ? user
        : formatUserOption(user);

    if (existingIds.has(option.value)) {
      return;
    }
    existingIds.add(option.value);
    merged.push(option);
  });

  return merged;
};

export const createEmptyMultiplierRule = () => ({
  localKey: nextMultiplierRuleKey(),
  scope: 'all',
  group_names: [],
  user_ids: [],
  metrics: ['daily'],
  multiplier: 1.1,
  userSearchKeyword: '',
});

export const parseMultiplierRules = (rawValue) => {
  let parsed = [];
  try {
    parsed = rawValue ? JSON.parse(rawValue) : [];
  } catch {
    parsed = [];
  }

  if (!Array.isArray(parsed)) {
    return [];
  }

  return parsed.map((rule) => ({
    localKey: nextMultiplierRuleKey(),
    scope: ['all', 'groups', 'users'].includes(rule?.scope) ? rule.scope : 'all',
    group_names: normalizeStringList(rule?.group_names),
    user_ids: normalizeNumberList(rule?.user_ids),
    metrics: normalizeStringList(rule?.metrics),
    multiplier:
      typeof rule?.multiplier === 'number' && Number.isFinite(rule.multiplier)
        ? rule.multiplier
        : 1,
    userSearchKeyword: '',
  }));
};

export const normalizeMultiplierRulesForSave = (rules = []) =>
  rules.map((rule) => {
    const normalizedRule = {
      scope: rule.scope,
      metrics: normalizeStringList(rule.metrics),
      multiplier: Number(rule.multiplier),
    };

    if (rule.scope === 'groups') {
      normalizedRule.group_names = normalizeStringList(rule.group_names);
    }
    if (rule.scope === 'users') {
      normalizedRule.user_ids = normalizeNumberList(rule.user_ids);
    }

    return normalizedRule;
  });
