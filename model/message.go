package model

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MessageStatusDraft  = "draft"
	MessageStatusOnline = "online"

	MessageSourceAdmin  = "admin"
	MessageSourceSystem = "system"

	MessageTargetAll    = "all"
	MessageTargetGroups = "groups"
	MessageTargetUsers  = "users"
)

type Message struct {
	Id            uint       `json:"id"`
	Title         string     `json:"title" gorm:"size:255;not null"`
	Content       string     `json:"content" gorm:"type:text;not null"`
	Status        string     `json:"status" gorm:"size:16;index;not null;default:'draft'"`
	Source        string     `json:"source" gorm:"size:16;index;not null;default:'admin'"`
	TargetType    string     `json:"target_type" gorm:"size:16;index;not null;default:'all'"`
	TargetGroups  string     `json:"target_groups" gorm:"type:text"`
	TargetUserIds string     `json:"target_user_ids" gorm:"type:text"`
	PublishedAt   *time.Time `json:"published_at,omitempty" gorm:"index"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UserMessage struct {
	Id          uint       `json:"id"`
	UserId      int        `json:"user_id" gorm:"index:idx_user_message,unique;index"`
	MessageId   uint       `json:"message_id" gorm:"index:idx_user_message,unique;index"`
	ReadAt      *time.Time `json:"read_at,omitempty" gorm:"index"`
	EmailSentAt *time.Time `json:"email_sent_at,omitempty"`
	EmailError  string     `json:"email_error,omitempty" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Message     Message    `json:"message" gorm:"foreignKey:MessageId;constraint:OnDelete:CASCADE"`
}

type MessageRecipient struct {
	UserId      int
	CAHID       string `gorm:"column:cah_id"`
	Username    string
	DisplayName string `gorm:"column:display_name"`
	Email       string
	Group       string `gorm:"column:group_name"`
}

type MessageTargetUserOption struct {
	Id          int    `json:"id"`
	CAHID       string `json:"cah_id,omitempty" gorm:"column:cah_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name,omitempty" gorm:"column:display_name"`
	Email       string `json:"email,omitempty"`
	Group       string `json:"group,omitempty" gorm:"column:group_name"`
}

type UserMessageView struct {
	Id          uint       `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	Source      string     `json:"source"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	EmailSentAt *time.Time `json:"email_sent_at,omitempty"`
}

func GetAdminMessages(pageInfo *common.PageInfo) ([]*Message, int64, error) {
	var (
		messages []*Message
		total    int64
	)

	query := DB.Model(&Message{}).Where("source = ?", MessageSourceAdmin)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("CASE WHEN published_at IS NULL THEN created_at ELSE published_at END DESC").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Find(&messages).Error; err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

func GetMessageByID(id uint) (*Message, error) {
	var message Message
	if err := DB.First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func CreateMessage(message *Message) error {
	return DB.Create(message).Error
}

func SaveMessage(message *Message) error {
	return DB.Save(message).Error
}

func DeleteDraftMessage(id uint) error {
	result := DB.Where("id = ? AND status = ? AND source = ?", id, MessageStatusDraft, MessageSourceAdmin).Delete(&Message{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (message *Message) GetTargetType() string {
	targetType := strings.TrimSpace(message.TargetType)
	if targetType == "" {
		return MessageTargetAll
	}
	return targetType
}

func (message *Message) GetTargetGroups() []string {
	if strings.TrimSpace(message.TargetGroups) == "" {
		return nil
	}
	var groups []string
	if err := common.UnmarshalJsonStr(message.TargetGroups, &groups); err != nil {
		return nil
	}
	return normalizeTargetGroups(groups)
}

func (message *Message) SetTargetGroups(groups []string) error {
	normalized := normalizeTargetGroups(groups)
	if len(normalized) == 0 {
		message.TargetGroups = ""
		return nil
	}
	bytes, err := common.Marshal(normalized)
	if err != nil {
		return err
	}
	message.TargetGroups = string(bytes)
	return nil
}

func (message *Message) GetTargetUserIDs() []int {
	if strings.TrimSpace(message.TargetUserIds) == "" {
		return nil
	}
	var userIDs []int
	if err := common.UnmarshalJsonStr(message.TargetUserIds, &userIDs); err != nil {
		return nil
	}
	return normalizeTargetUserIDs(userIDs)
}

func (message *Message) SetTargetUserIDs(userIDs []int) error {
	normalized := normalizeTargetUserIDs(userIDs)
	if len(normalized) == 0 {
		message.TargetUserIds = ""
		return nil
	}
	bytes, err := common.Marshal(normalized)
	if err != nil {
		return err
	}
	message.TargetUserIds = string(bytes)
	return nil
}

func ListEnabledUserRecipients() ([]MessageRecipient, error) {
	recipients := make([]MessageRecipient, 0)
	err := DB.Model(&User{}).
		Where("status = ?", common.UserStatusEnabled).
		Select(fmt.Sprintf("id as user_id, cah_id, username, display_name, email, %s as group_name", commonGroupCol)).
		Find(&recipients).Error
	return recipients, err
}

func ListEnabledUserRecipientsByScope(targetType string, groups []string, userIDs []int) ([]MessageRecipient, error) {
	recipients := make([]MessageRecipient, 0)
	query := DB.Model(&User{}).
		Where("status = ?", common.UserStatusEnabled)

	switch strings.TrimSpace(targetType) {
	case "", MessageTargetAll:
		// all enabled users
	case MessageTargetGroups:
		normalizedGroups := normalizeTargetGroups(groups)
		if len(normalizedGroups) == 0 {
			return nil, errors.New("未指定有效分组")
		}
		query = query.Where(commonGroupCol+" IN ?", normalizedGroups)
	case MessageTargetUsers:
		normalizedUserIDs := normalizeTargetUserIDs(userIDs)
		if len(normalizedUserIDs) == 0 {
			return nil, errors.New("未指定有效用户")
		}
		query = query.Where("id IN ?", normalizedUserIDs)
	default:
		return nil, errors.New("无效的消息发送范围")
	}

	err := query.Select(fmt.Sprintf("id as user_id, cah_id, username, display_name, email, %s as group_name", commonGroupCol)).
		Order("id desc").
		Find(&recipients).Error
	return recipients, err
}

func CreateUserMessageDeliveries(messageID uint, recipients []MessageRecipient) error {
	if len(recipients) == 0 {
		return nil
	}

	deliveries := make([]UserMessage, 0, len(recipients))
	for _, recipient := range recipients {
		deliveries = append(deliveries, UserMessage{
			UserId:    recipient.UserId,
			MessageId: messageID,
		})
	}

	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "message_id"}},
		DoNothing: true,
	}).CreateInBatches(deliveries, 200).Error
}

func CreateSingleUserMessageDelivery(messageID uint, userID int) error {
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "message_id"}},
		DoNothing: true,
	}).Create(&UserMessage{
		UserId:    userID,
		MessageId: messageID,
	}).Error
}

func GetUserMessages(userID int, pageInfo *common.PageInfo) ([]*UserMessageView, int64, error) {
	var (
		items []*UserMessageView
		total int64
	)

	query := DB.Table("user_messages").
		Joins("JOIN messages ON messages.id = user_messages.message_id").
		Where("user_messages.user_id = ? AND messages.status = ?", userID, MessageStatusOnline)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Select("messages.id, messages.title, messages.content, messages.status, messages.source, messages.published_at, messages.created_at, user_messages.read_at, user_messages.email_sent_at").
		Order("messages.published_at DESC, messages.id DESC").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func CountUnreadUserMessages(userID int) (int64, error) {
	var total int64
	err := DB.Table("user_messages").
		Joins("JOIN messages ON messages.id = user_messages.message_id").
		Where("user_messages.user_id = ? AND user_messages.read_at IS NULL AND messages.status = ?", userID, MessageStatusOnline).
		Count(&total).Error
	return total, err
}

func MarkUserMessageRead(userID int, messageID uint) error {
	now := time.Now()
	result := DB.Model(&UserMessage{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Where("read_at IS NULL").
		Update("read_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var exists int64
		if err := DB.Model(&UserMessage{}).Where("user_id = ? AND message_id = ?", userID, messageID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}

func UpdateUserMessageEmailStatus(userID int, messageID uint, emailSentAt *time.Time, emailError string) error {
	return DB.Model(&UserMessage{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Updates(map[string]interface{}{
			"email_sent_at": emailSentAt,
			"email_error":   emailError,
		}).Error
}

func ListFailedEmailRecipientsForMessage(messageID uint) ([]MessageRecipient, error) {
	recipients := make([]MessageRecipient, 0)
	err := DB.Table("user_messages").
		Joins("JOIN users ON users.id = user_messages.user_id").
		Where("user_messages.message_id = ? AND user_messages.email_sent_at IS NULL AND user_messages.email_error <> ''", messageID).
		Where("users.status = ?", common.UserStatusEnabled).
		Where("users.email <> ''").
		Select(fmt.Sprintf("users.id as user_id, users.cah_id, users.username, users.display_name, users.email, %s as group_name", commonGroupCol)).
		Find(&recipients).Error
	return recipients, err
}

func GetMessageDeliveryStats(messageIDs []uint) (map[uint]map[string]int64, error) {
	stats := make(map[uint]map[string]int64)
	if len(messageIDs) == 0 {
		return stats, nil
	}

	type statRow struct {
		MessageId   uint
		Total       int64
		ReadTotal   int64
		EmailSent   int64
		EmailFailed int64
	}

	rows := make([]statRow, 0, len(messageIDs))
	err := DB.Model(&UserMessage{}).
		Select("message_id, COUNT(*) AS total, SUM(CASE WHEN read_at IS NOT NULL THEN 1 ELSE 0 END) AS read_total, SUM(CASE WHEN email_sent_at IS NOT NULL THEN 1 ELSE 0 END) AS email_sent, SUM(CASE WHEN email_error <> '' AND email_sent_at IS NULL THEN 1 ELSE 0 END) AS email_failed").
		Where("message_id IN ?", messageIDs).
		Group("message_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		stats[row.MessageId] = map[string]int64{
			"total":        row.Total,
			"read_total":   row.ReadTotal,
			"email_sent":   row.EmailSent,
			"email_failed": row.EmailFailed,
		}
	}
	return stats, nil
}

func GetMessageTargetUserOptions(userIDs []int) ([]MessageTargetUserOption, error) {
	normalizedUserIDs := normalizeTargetUserIDs(userIDs)
	if len(normalizedUserIDs) == 0 {
		return []MessageTargetUserOption{}, nil
	}

	options := make([]MessageTargetUserOption, 0, len(normalizedUserIDs))
	if err := DB.Model(&User{}).
		Select(fmt.Sprintf("id, cah_id, username, display_name, email, %s as group_name", commonGroupCol)).
		Where("id IN ?", normalizedUserIDs).
		Scan(&options).Error; err != nil {
		return nil, err
	}

	order := make(map[int]int, len(normalizedUserIDs))
	for idx, userID := range normalizedUserIDs {
		order[userID] = idx
	}
	slices.SortFunc(options, func(a, b MessageTargetUserOption) int {
		return order[a.Id] - order[b.Id]
	})
	return options, nil
}

func MessageExistsBySignature(title string, content string, publishedAt time.Time, source string) (bool, error) {
	var count int64
	err := DB.Model(&Message{}).
		Where("title = ? AND content = ? AND source = ? AND published_at = ?", title, content, source, publishedAt).
		Count(&count).Error
	return count > 0, err
}

func GetUserByEmailAddress(email string) (*User, error) {
	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func EnsureMessageEditable(message *Message) error {
	if message == nil {
		return errors.New("消息不存在")
	}
	if message.Source != MessageSourceAdmin {
		return errors.New("系统消息不支持编辑")
	}
	if message.Status != MessageStatusDraft {
		return errors.New("已上线的消息不可编辑")
	}
	return nil
}

func MarkUserMessagesRead(userID int, messageIDs []uint) (int64, error) {
	if len(messageIDs) == 0 {
		return 0, nil
	}
	now := time.Now()
	result := DB.Model(&UserMessage{}).
		Where("user_id = ? AND message_id IN ?", userID, messageIDs).
		Where("read_at IS NULL").
		Update("read_at", &now)
	return result.RowsAffected, result.Error
}

func normalizeTargetGroups(groups []string) []string {
	normalized := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		if _, exists := seen[group]; exists {
			continue
		}
		seen[group] = struct{}{}
		normalized = append(normalized, group)
	}
	return normalized
}

func normalizeTargetUserIDs(userIDs []int) []int {
	normalized := make([]int, 0, len(userIDs))
	seen := make(map[int]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		normalized = append(normalized, userID)
	}
	return normalized
}
