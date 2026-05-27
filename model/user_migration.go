package model

const (
	UserMigrationStatusDraft     = "draft"
	UserMigrationStatusActive    = "active"
	UserMigrationStatusClosed    = "closed"
	UserMigrationStatusCancelled = "cancelled"

	UserMigrationTargetStatusPending     = "pending"
	UserMigrationTargetStatusEmailSent   = "email_sent"
	UserMigrationTargetStatusEmailFailed = "email_failed"
	UserMigrationTargetStatusOpened      = "opened"
	UserMigrationTargetStatusCaptured    = "captured"
	UserMigrationTargetStatusMigrated    = "migrated"

	UserMigrationImportStatusPendingSetup = "pending_setup"
	UserMigrationImportStatusActive       = "active"
	UserMigrationImportStatusEmailFailed  = "email_failed"
)

type UserMigration struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	MigrateID      string `json:"migrate_id" gorm:"size:32;uniqueIndex;not null"`
	Name           string `json:"name" gorm:"size:120;not null"`
	Description    string `json:"description" gorm:"type:text"`
	Expression     string `json:"expression" gorm:"type:text"`
	Status         string `json:"status" gorm:"size:24;index;not null;default:'draft'"`
	CreatedBy      int    `json:"created_by" gorm:"index"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
	ClosedAt       int64  `json:"closed_at" gorm:"bigint;index"`
	TargetCount    int    `json:"target_count" gorm:"default:0"`
	EmailSentCount int    `json:"email_sent_count" gorm:"default:0"`
	CapturedCount  int    `json:"captured_count" gorm:"default:0"`
	MigratedCount  int    `json:"migrated_count" gorm:"default:0"`
}

type UserMigrationTarget struct {
	ID                 uint   `json:"id" gorm:"primaryKey"`
	MigrateID          string `json:"migrate_id" gorm:"size:32;index;uniqueIndex:idx_migration_target_user;not null"`
	UserID             int    `json:"user_id" gorm:"index;uniqueIndex:idx_migration_target_user;not null"`
	CAHID              string `json:"cah_id" gorm:"column:cah_id;size:16;index"`
	Email              string `json:"email" gorm:"size:255;index"`
	Status             string `json:"status" gorm:"size:24;index;not null;default:'pending'"`
	MigrationTokenHash string `json:"-" gorm:"size:64;index;not null"`
	AccessTokenHash    string `json:"-" gorm:"size:64;not null"`
	UserTokenHash      string `json:"-" gorm:"size:64"`
	EmailSentAt        int64  `json:"email_sent_at" gorm:"bigint;index"`
	EmailError         string `json:"email_error" gorm:"type:text"`
	OpenedAt           int64  `json:"opened_at" gorm:"bigint;index"`
	CapturedAt         int64  `json:"captured_at" gorm:"bigint;index"`
	LastExportedAt     int64  `json:"last_exported_at" gorm:"bigint;index"`
	MigratedAt         int64  `json:"migrated_at" gorm:"bigint;index"`
	DataJSON           string `json:"data_json,omitempty" gorm:"type:longtext"`
	CreatedAt          int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt          int64  `json:"updated_at" gorm:"bigint"`
}

type UserMigrationImport struct {
	ID               uint   `json:"id" gorm:"primaryKey"`
	ImportID         string `json:"import_id" gorm:"size:32;uniqueIndex;not null"`
	CAHID            string `json:"cah_id" gorm:"column:cah_id;size:16;index;not null"`
	Email            string `json:"email" gorm:"size:255;index;not null"`
	UserID           int    `json:"user_id" gorm:"index"`
	Status           string `json:"status" gorm:"size:24;index;not null;default:'pending_setup'"`
	SetupTokenHash   string `json:"-" gorm:"size:64;index;not null"`
	AccessTokenHash  string `json:"-" gorm:"size:64;not null"`
	DataJSON         string `json:"data_json,omitempty" gorm:"type:longtext"`
	CreatedBy        int    `json:"created_by" gorm:"index"`
	CreatedAt        int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt        int64  `json:"updated_at" gorm:"bigint"`
	ImportedAt       int64  `json:"imported_at" gorm:"bigint;index"`
	EmailSentAt      int64  `json:"email_sent_at" gorm:"bigint;index"`
	EmailError       string `json:"email_error" gorm:"type:text"`
	SetupCompletedAt int64  `json:"setup_completed_at" gorm:"bigint;index"`
}
