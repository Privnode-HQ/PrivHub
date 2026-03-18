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
import { Avatar, Typography } from '@douyinfe/semi-ui';
import { TicketPercent } from 'lucide-react';
import CardPro from '../../common/ui/CardPro';
import CompactModeToggle from '../../common/ui/CompactModeToggle';
import TopupCouponsTable from './TopupCouponsTable';
import TopupCouponsActions from './TopupCouponsActions';
import TopupCouponsFilters from './TopupCouponsFilters';
import EditTopupCouponModal from './modals/EditTopupCouponModal';
import RevokeTopupCouponModal from './modals/RevokeTopupCouponModal';
import { useTopupCouponsData } from '../../../hooks/topupCoupons/useTopupCouponsData';
import { useIsMobile } from '../../../hooks/common/useIsMobile';
import { createCardProPagination } from '../../../helpers/utils';

const { Text } = Typography;

const TopupCouponsPage = () => {
  const topupCouponsData = useTopupCouponsData();
  const isMobile = useIsMobile();

  return (
    <>
      <EditTopupCouponModal
        refresh={topupCouponsData.refresh}
        editingCoupon={topupCouponsData.editingCoupon}
        visible={topupCouponsData.showEdit}
        handleClose={topupCouponsData.closeEdit}
      />

      <RevokeTopupCouponModal
        visible={topupCouponsData.showRevoke}
        coupon={topupCouponsData.revokingCoupon}
        onCancel={topupCouponsData.closeRevoke}
        onConfirm={topupCouponsData.revokeCoupon}
        loading={topupCouponsData.loading}
        t={topupCouponsData.t}
      />

      <CardPro
        type='type1'
        descriptionArea={
          <div className='flex flex-col md:flex-row justify-between items-start md:items-center gap-2 w-full'>
            <div className='flex items-center text-emerald-500'>
              <Avatar size='small' color='green' className='mr-2 shadow-md'>
                <TicketPercent size={14} />
              </Avatar>
              <Text>{topupCouponsData.t('折扣中心')}</Text>
            </div>

            <CompactModeToggle
              compactMode={topupCouponsData.compactMode}
              setCompactMode={topupCouponsData.setCompactMode}
              t={topupCouponsData.t}
            />
          </div>
        }
        actionsArea={
          <div className='flex flex-col md:flex-row justify-between items-center gap-2 w-full'>
            <TopupCouponsActions
              openCreate={topupCouponsData.openCreate}
              refresh={topupCouponsData.refresh}
              t={topupCouponsData.t}
            />

            <div className='w-full md:w-full lg:w-auto order-1 md:order-2'>
              <TopupCouponsFilters
                formInitValues={topupCouponsData.formInitValues}
                setFormApi={topupCouponsData.setFormApi}
                searchTopupCoupons={topupCouponsData.searchTopupCoupons}
                loading={topupCouponsData.loading}
                searching={topupCouponsData.searching}
                t={topupCouponsData.t}
              />
            </div>
          </div>
        }
        paginationArea={createCardProPagination({
          currentPage: topupCouponsData.activePage,
          pageSize: topupCouponsData.pageSize,
          total: topupCouponsData.tokenCount,
          onPageChange: topupCouponsData.handlePageChange,
          onPageSizeChange: topupCouponsData.handlePageSizeChange,
          isMobile: isMobile,
          t: topupCouponsData.t,
        })}
        t={topupCouponsData.t}
      >
        <TopupCouponsTable {...topupCouponsData} />
      </CardPro>
    </>
  );
};

export default TopupCouponsPage;
