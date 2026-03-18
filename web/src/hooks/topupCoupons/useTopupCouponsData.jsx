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

import { useEffect, useState } from 'react';
import { API, showError, showSuccess } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTranslation } from 'react-i18next';
import { useTableCompactMode } from '../common/useTableCompactMode';

const defaultEditingCoupon = {
  id: undefined,
};

export const useTopupCouponsData = () => {
  const { t } = useTranslation();
  const [coupons, setCoupons] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [tokenCount, setTokenCount] = useState(0);
  const [editingCoupon, setEditingCoupon] = useState(defaultEditingCoupon);
  const [showEdit, setShowEdit] = useState(false);
  const [revokingCoupon, setRevokingCoupon] = useState(null);
  const [showRevoke, setShowRevoke] = useState(false);
  const [formApi, setFormApi] = useState(null);
  const [compactMode, setCompactMode] = useTableCompactMode('topup-coupons');

  const formInitValues = {
    searchKeyword: '',
    status: '',
    userId: undefined,
  };

  const getFormValues = () => {
    const values = formApi ? formApi.getValues() : {};
    return {
      searchKeyword: values.searchKeyword || '',
      status: values.status || '',
      userId: values.userId || '',
    };
  };

  const buildQuery = (page = 1, size = pageSize) => {
    const { searchKeyword, status, userId } = getFormValues();
    const params = new URLSearchParams({
      p: String(page),
      page_size: String(size),
    });

    if (searchKeyword) {
      params.set('keyword', searchKeyword);
    }
    if (status) {
      params.set('status', status);
    }
    if (userId) {
      params.set('user_id', String(userId));
    }
    return params.toString();
  };

  const loadTopupCoupons = async (page = 1, size = pageSize) => {
    setLoading(true);
    try {
      const res = await API.get(`/api/topup-coupon/?${buildQuery(page, size)}`);
      const { success, message, data } = res.data;
      if (success) {
        setCoupons(data.items || []);
        setActivePage(data.page || 1);
        setTokenCount(data.total || 0);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const searchTopupCoupons = async () => {
    setSearching(true);
    await loadTopupCoupons(1, pageSize);
    setSearching(false);
  };

  const refresh = async (page = activePage) => {
    await loadTopupCoupons(page, pageSize);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    loadTopupCoupons(page, pageSize);
  };

  const handlePageSizeChange = (size) => {
    setPageSize(size);
    setActivePage(1);
    loadTopupCoupons(1, size);
  };

  const openCreate = () => {
    setEditingCoupon(defaultEditingCoupon);
    setShowEdit(true);
  };

  const openEdit = (coupon) => {
    setEditingCoupon(coupon);
    setShowEdit(true);
  };

  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingCoupon(defaultEditingCoupon);
    }, 300);
  };

  const openRevoke = (coupon) => {
    setRevokingCoupon(coupon);
    setShowRevoke(true);
  };

  const closeRevoke = () => {
    setShowRevoke(false);
    setTimeout(() => {
      setRevokingCoupon(null);
    }, 300);
  };

  const revokeCoupon = async (reason) => {
    if (!revokingCoupon?.id) return;
    setLoading(true);
    try {
      const res = await API.put('/api/topup-coupon/', {
        id: revokingCoupon.id,
        action: 'revoke',
        revoke_reason: reason,
      });
      const { success, message } = res.data;
      if (success) {
        showSuccess(t('优惠券已撤销'));
        closeRevoke();
        await refresh();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const handleRow = (record) => {
    if (record.status !== 'available') {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    }
    return {};
  };

  useEffect(() => {
    loadTopupCoupons(1, pageSize).catch(showError);
  }, [pageSize]);

  return {
    coupons,
    loading,
    searching,
    activePage,
    pageSize,
    tokenCount,
    editingCoupon,
    showEdit,
    revokingCoupon,
    showRevoke,
    formInitValues,
    compactMode,
    setCompactMode,
    setFormApi,
    searchTopupCoupons,
    refresh,
    handlePageChange,
    handlePageSizeChange,
    handleRow,
    openCreate,
    openEdit,
    closeEdit,
    openRevoke,
    closeRevoke,
    revokeCoupon,
    t,
  };
};
