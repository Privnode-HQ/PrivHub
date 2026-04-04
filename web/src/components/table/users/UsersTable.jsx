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

import React, { useMemo, useState } from 'react';
import { Empty, Modal } from '@douyinfe/semi-ui';
import CardTable from '../../common/ui/CardTable';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { getUsersColumns } from './UsersColumnDefs';
import PromoteUserModal from './modals/PromoteUserModal';
import DemoteUserModal from './modals/DemoteUserModal';
import EnableDisableUserModal from './modals/EnableDisableUserModal';
import DeleteUserModal from './modals/DeleteUserModal';
import ResetPasskeyModal from './modals/ResetPasskeyModal';
import ResetTwoFAModal from './modals/ResetTwoFAModal';
import ImpersonationUserModal from './modals/ImpersonationUserModal';

const UsersTable = (usersData) => {
  const {
    users,
    loading,
    activePage,
    pageSize,
    userCount,
    compactMode,
    handlePageChange,
    handlePageSizeChange,
    handleRow,
    setEditingUser,
    setShowEditUser,
    manageUser,
    refresh,
    resetUserPasskey,
    resetUserTwoFA,
    startImpersonation,
    impersonationLoading,
    readOnlyAdmin,
    t,
  } = usersData;

  // Modal states
  const [showPromoteModal, setShowPromoteModal] = useState(false);
  const [showDemoteModal, setShowDemoteModal] = useState(false);
  const [showEnableDisableModal, setShowEnableDisableModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [modalUser, setModalUser] = useState(null);
  const [enableDisableAction, setEnableDisableAction] = useState('');
  const [showResetPasskeyModal, setShowResetPasskeyModal] = useState(false);
  const [showResetTwoFAModal, setShowResetTwoFAModal] = useState(false);
  const [showImpersonationModal, setShowImpersonationModal] = useState(false);

  // Modal handlers
  const showPromoteUserModal = (user) => {
    setModalUser(user);
    setShowPromoteModal(true);
  };

  const showDemoteUserModal = (user) => {
    setModalUser(user);
    setShowDemoteModal(true);
  };

  const showEnableDisableUserModal = (user, action) => {
    setModalUser(user);
    setEnableDisableAction(action);
    setShowEnableDisableModal(true);
  };

  const showDeleteUserModal = (user) => {
    setModalUser(user);
    setShowDeleteModal(true);
  };

  const showResetPasskeyUserModal = (user) => {
    setModalUser(user);
    setShowResetPasskeyModal(true);
  };

  const showResetTwoFAUserModal = (user) => {
    setModalUser(user);
    setShowResetTwoFAModal(true);
  };

  const showImpersonationUserModal = (user) => {
    setModalUser(user);
    setShowImpersonationModal(true);
  };

  const showForceLogoutUserConfirm = (user) => {
    Modal.confirm({
      title: t('强制退出该用户登录'),
      content: t('该用户当前所有 Web 登录会话会立即失效，需要重新登录。'),
      okText: t('确认退出'),
      cancelText: t('取消'),
      onOk: () => manageUser(user.id, 'logout', user),
    });
  };

  const showRequirePasswordResetUserConfirm = (user) => {
    Modal.confirm({
      title: t('要求用户修改密码'),
      content: t('用户需要先完成密码修改，才能继续访问其他功能。'),
      okText: t('确认要求'),
      cancelText: t('取消'),
      onOk: () => manageUser(user.id, 'require_password_reset', user),
    });
  };

  const showRequireEmailBindUserConfirm = (user) => {
    Modal.confirm({
      title: t('要求用户绑定邮箱'),
      content: t('用户需要先完成邮箱绑定，才能继续访问其他功能。'),
      okText: t('确认要求'),
      cancelText: t('取消'),
      onOk: () => manageUser(user.id, 'require_email_bind', user),
    });
  };

  // Modal confirm handlers
  const handlePromoteConfirm = () => {
    manageUser(modalUser.id, 'promote', modalUser);
    setShowPromoteModal(false);
  };

  const handleDemoteConfirm = () => {
    manageUser(modalUser.id, 'demote', modalUser);
    setShowDemoteModal(false);
  };

  const handleEnableDisableConfirm = (banReason) => {
    manageUser(modalUser.id, enableDisableAction, modalUser, banReason);
    setShowEnableDisableModal(false);
  };

  const handleResetPasskeyConfirm = async () => {
    await resetUserPasskey(modalUser);
    setShowResetPasskeyModal(false);
  };

  const handleResetTwoFAConfirm = async () => {
    await resetUserTwoFA(modalUser);
    setShowResetTwoFAModal(false);
  };

  // Get all columns
  const columns = useMemo(() => {
    return getUsersColumns({
      t,
      setEditingUser,
      setShowEditUser,
      showPromoteModal: showPromoteUserModal,
      showDemoteModal: showDemoteUserModal,
      showEnableDisableModal: showEnableDisableUserModal,
      showDeleteModal: showDeleteUserModal,
      showResetPasskeyModal: showResetPasskeyUserModal,
      showResetTwoFAModal: showResetTwoFAUserModal,
      showImpersonationModal: showImpersonationUserModal,
      showForceLogoutConfirm: showForceLogoutUserConfirm,
      showRequirePasswordResetConfirm: showRequirePasswordResetUserConfirm,
      showRequireEmailBindConfirm: showRequireEmailBindUserConfirm,
      readOnlyAdmin,
    });
  }, [
    t,
    setEditingUser,
    setShowEditUser,
    showPromoteUserModal,
    showDemoteUserModal,
    showEnableDisableUserModal,
    showDeleteUserModal,
    showResetPasskeyUserModal,
    showResetTwoFAUserModal,
    showImpersonationUserModal,
    showForceLogoutUserConfirm,
    showRequirePasswordResetUserConfirm,
    showRequireEmailBindUserConfirm,
    readOnlyAdmin,
  ]);

  // Handle compact mode by removing fixed positioning
  const tableColumns = useMemo(() => {
    return compactMode
      ? columns.map((col) => {
          if (col.dataIndex === 'operate') {
            const { fixed, ...rest } = col;
            return rest;
          }
          return col;
        })
      : columns;
  }, [compactMode, columns]);

  return (
    <>
      <CardTable
        columns={tableColumns}
        dataSource={users}
        scroll={compactMode ? undefined : { x: 'max-content' }}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: userCount,
          pageSizeOpts: [10, 20, 50, 100],
          showSizeChanger: true,
          onPageSizeChange: handlePageSizeChange,
          onPageChange: handlePageChange,
        }}
        hidePagination={true}
        loading={loading}
        onRow={handleRow}
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
            }
            description={t('搜索无结果')}
            style={{ padding: 30 }}
          />
        }
        className='overflow-hidden'
        size='middle'
      />

      {/* Modal components */}
      <PromoteUserModal
        visible={showPromoteModal}
        onCancel={() => setShowPromoteModal(false)}
        onConfirm={handlePromoteConfirm}
        user={modalUser}
        t={t}
      />

      <DemoteUserModal
        visible={showDemoteModal}
        onCancel={() => setShowDemoteModal(false)}
        onConfirm={handleDemoteConfirm}
        user={modalUser}
        t={t}
      />

      <EnableDisableUserModal
        visible={showEnableDisableModal}
        onCancel={() => setShowEnableDisableModal(false)}
        onConfirm={handleEnableDisableConfirm}
        user={modalUser}
        action={enableDisableAction}
        t={t}
      />

      <DeleteUserModal
        visible={showDeleteModal}
        onCancel={() => setShowDeleteModal(false)}
        user={modalUser}
        users={users}
        activePage={activePage}
        refresh={refresh}
        manageUser={manageUser}
        t={t}
      />

      <ResetPasskeyModal
        visible={showResetPasskeyModal}
        onCancel={() => setShowResetPasskeyModal(false)}
        onConfirm={handleResetPasskeyConfirm}
        user={modalUser}
        t={t}
      />

      <ResetTwoFAModal
        visible={showResetTwoFAModal}
        onCancel={() => setShowResetTwoFAModal(false)}
        onConfirm={handleResetTwoFAConfirm}
        user={modalUser}
        t={t}
      />

      <ImpersonationUserModal
        visible={showImpersonationModal}
        onCancel={() => setShowImpersonationModal(false)}
        onSubmit={async (options) => {
          const success = await startImpersonation(modalUser, options);
          if (success) {
            setShowImpersonationModal(false);
          }
        }}
        user={modalUser}
        loading={impersonationLoading}
        t={t}
      />
    </>
  );
};

export default UsersTable;
