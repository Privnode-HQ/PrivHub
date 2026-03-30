package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/console_setting"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminMessageDraftInput struct {
	Title         string
	Content       string
	TargetType    string
	TargetGroups  []string
	TargetUserIDs []int
}

func CreateAdminDraftMessage(input AdminMessageDraftInput) (*model.Message, error) {
	if err := validateAdminMessageDraftInput(&input); err != nil {
		return nil, err
	}

	message := &model.Message{
		Title:      strings.TrimSpace(input.Title),
		Content:    strings.TrimSpace(input.Content),
		Status:     model.MessageStatusDraft,
		Source:     model.MessageSourceAdmin,
		TargetType: normalizeMessageTargetType(input.TargetType),
	}
	if err := message.SetTargetGroups(input.TargetGroups); err != nil {
		return nil, err
	}
	if err := message.SetTargetUserIDs(input.TargetUserIDs); err != nil {
		return nil, err
	}
	if err := model.CreateMessage(message); err != nil {
		return nil, err
	}
	return message, nil
}

func UpdateAdminDraftMessage(id uint, input AdminMessageDraftInput) (*model.Message, error) {
	if err := validateAdminMessageDraftInput(&input); err != nil {
		return nil, err
	}

	message, err := model.GetMessageByID(id)
	if err != nil {
		return nil, err
	}
	if err = model.EnsureMessageEditable(message); err != nil {
		return nil, err
	}

	message.Title = strings.TrimSpace(input.Title)
	message.Content = strings.TrimSpace(input.Content)
	message.TargetType = normalizeMessageTargetType(input.TargetType)
	if err = message.SetTargetGroups(input.TargetGroups); err != nil {
		return nil, err
	}
	if err = message.SetTargetUserIDs(input.TargetUserIDs); err != nil {
		return nil, err
	}
	if err = model.SaveMessage(message); err != nil {
		return nil, err
	}
	return message, nil
}

func PublishAdminMessage(id uint) (*model.Message, error) {
	message, err := model.GetMessageByID(id)
	if err != nil {
		return nil, err
	}
	if err = model.EnsureMessageEditable(message); err != nil {
		return nil, err
	}

	recipients, err := model.ListEnabledUserRecipientsByScope(
		message.GetTargetType(),
		message.GetTargetGroups(),
		message.GetTargetUserIDs(),
	)
	if err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		return nil, errors.New("没有匹配到可投递的用户")
	}

	now := time.Now()
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		message.Status = model.MessageStatusOnline
		message.PublishedAt = &now
		if err := tx.Save(message).Error; err != nil {
			return err
		}

		deliveries := make([]model.UserMessage, 0, len(recipients))
		for _, recipient := range recipients {
			deliveries = append(deliveries, model.UserMessage{
				UserId:    recipient.UserId,
				MessageId: message.Id,
			})
		}
		return tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).CreateInBatches(deliveries, 200).Error
	})
	if err != nil {
		return nil, err
	}

	dispatchMessageEmailsAsync(*message, recipients, common.GenerateEmailIdempotencyKey("message-delivery-batch", fmt.Sprintf("%d", message.Id), now.Format(time.RFC3339Nano)))
	return message, nil
}

func CopyAdminMessage(id uint) (*model.Message, error) {
	source, err := model.GetMessageByID(id)
	if err != nil {
		return nil, err
	}
	if source.Source != model.MessageSourceAdmin {
		return nil, errors.New("仅支持复制后台消息")
	}

	input := AdminMessageDraftInput{
		Title:         source.Title,
		Content:       source.Content,
		TargetType:    source.GetTargetType(),
		TargetGroups:  source.GetTargetGroups(),
		TargetUserIDs: source.GetTargetUserIDs(),
	}
	return CreateAdminDraftMessage(input)
}

func DeliverSystemMessageToUser(userID int, cahID string, username string, displayName string, email string, title string, content string) error {
	now := time.Now()
	message := &model.Message{
		Title:       strings.TrimSpace(title),
		Content:     strings.TrimSpace(content),
		Status:      model.MessageStatusOnline,
		Source:      model.MessageSourceSystem,
		TargetType:  model.MessageTargetUsers,
		PublishedAt: &now,
	}
	if err := message.SetTargetUserIDs([]int{userID}); err != nil {
		return err
	}
	if err := model.CreateMessage(message); err != nil {
		return err
	}
	if err := model.CreateSingleUserMessageDelivery(message.Id, userID); err != nil {
		return err
	}

	if strings.TrimSpace(email) != "" {
		dispatchSingleMessageEmailAsync(*message, model.MessageRecipient{
			UserId:      userID,
			CAHID:       cahID,
			Username:    username,
			DisplayName: displayName,
			Email:       email,
		}, common.GenerateEmailIdempotencyKey("message-delivery", fmt.Sprintf("%d", message.Id), fmt.Sprintf("%d", userID), now.Format(time.RFC3339Nano)))
	}
	return nil
}

func RetryFailedMessageEmailDelivery(messageID uint) (int, error) {
	message, err := model.GetMessageByID(messageID)
	if err != nil {
		return 0, err
	}
	if message.Status != model.MessageStatusOnline {
		return 0, errors.New("仅支持重试已上线消息")
	}

	recipients, err := model.ListFailedEmailRecipientsForMessage(messageID)
	if err != nil {
		return 0, err
	}
	if len(recipients) == 0 {
		return 0, errors.New("没有可重试的失败邮件")
	}

	dispatchMessageEmailsAsync(*message, recipients, common.GenerateEmailIdempotencyKey("message-retry-batch", fmt.Sprintf("%d", message.Id), time.Now().Format(time.RFC3339Nano)))
	return len(recipients), nil
}

func dispatchMessageEmailsAsync(message model.Message, recipients []model.MessageRecipient, idempotencyKey string) {
	gopool.Go(func() {
		sendMessageEmailBatch(message, recipients, idempotencyKey)
	})
}

func dispatchSingleMessageEmailAsync(message model.Message, recipient model.MessageRecipient, idempotencyKey string) {
	gopool.Go(func() {
		sendMessageEmailBatch(message, []model.MessageRecipient{recipient}, idempotencyKey)
	})
}

func sendMessageEmailBatch(message model.Message, recipients []model.MessageRecipient, idempotencyKey string) {
	emailEntries := make([]common.BatchEmailEntry, 0, len(recipients))
	emailRecipients := make([]model.MessageRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		if strings.TrimSpace(recipient.Email) == "" {
			continue
		}
		trimmedEmail := strings.TrimSpace(recipient.Email)
		emailEntries = append(emailEntries, common.BatchEmailEntry{
			Recipient: trimmedEmail,
			Subject:   message.Title,
			Content:   message.Content,
			Context: common.EmailRecipientContext{
				Username:      recipient.Username,
				RecipientName: strings.TrimSpace(recipient.DisplayName),
				CAHID:         recipient.CAHID,
				Email:         trimmedEmail,
				PublishedAt:   messagePublishedAt(message),
			},
		})
		emailRecipients = append(emailRecipients, recipient)
	}
	if len(emailEntries) == 0 {
		return
	}

	results, err := common.SendBatchEmailsWithIdempotencyKey(emailEntries, idempotencyKey)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to send message email batch %d: %v", message.Id, err))
		for _, recipient := range recipients {
			if strings.TrimSpace(recipient.Email) == "" {
				continue
			}
			_ = model.UpdateUserMessageEmailStatus(recipient.UserId, message.Id, nil, err.Error())
		}
		return
	}

	for idx, result := range results {
		if idx >= len(emailRecipients) {
			continue
		}
		recipient := emailRecipients[idx]

		if result.Success {
			now := time.Now()
			_ = model.UpdateUserMessageEmailStatus(recipient.UserId, message.Id, &now, "")
			continue
		}

		errorMessage := strings.TrimSpace(result.Error)
		if errorMessage == "" {
			errorMessage = "邮件发送失败"
		}
		common.SysError(fmt.Sprintf("failed to send message email %d to user %d: %s", message.Id, recipient.UserId, errorMessage))
		_ = model.UpdateUserMessageEmailStatus(recipient.UserId, message.Id, nil, errorMessage)
	}
}

func messagePublishedAt(message model.Message) time.Time {
	if message.PublishedAt != nil {
		return *message.PublishedAt
	}
	return message.CreatedAt
}

func MigrateLegacyNoticesToMessages() {
	if hasMigratedLegacyMessages() {
		return
	}

	migrated := false
	if migrateLegacyAnnouncementOption() {
		migrated = true
	}
	if migrateLegacyNoticeOption() {
		migrated = true
	}

	if migrated {
		_ = model.UpdateOption("console_setting.announcements", "")
		_ = model.UpdateOption("console_setting.announcements_enabled", "false")
		_ = model.UpdateOption("Notice", "")
	}
	_ = model.UpdateOption("LegacyMessagesMigrated", "true")
}

func hasMigratedLegacyMessages() bool {
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	return strings.TrimSpace(common.OptionMap["LegacyMessagesMigrated"]) == "true"
}

func migrateLegacyAnnouncementOption() bool {
	raw := strings.TrimSpace(console_setting.GetConsoleSetting().Announcements)
	if raw == "" {
		return false
	}

	var items []struct {
		Content     string `json:"content"`
		PublishDate string `json:"publishDate"`
		Extra       string `json:"extra"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		common.SysError(fmt.Sprintf("failed to parse legacy announcements: %v", err))
		return false
	}

	recipients, err := model.ListEnabledUserRecipients()
	if err != nil {
		common.SysError(fmt.Sprintf("failed to list users for legacy announcement migration: %v", err))
		return false
	}

	migrated := false
	for _, item := range items {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}

		publishedAt, err := time.Parse(time.RFC3339, item.PublishDate)
		if err != nil {
			publishedAt = time.Now()
		}

		title := deriveLegacyMessageTitle(content, item.Extra, publishedAt)
		fullContent := content
		if strings.TrimSpace(item.Extra) != "" {
			fullContent = fmt.Sprintf("%s\n\n> %s", content, strings.TrimSpace(item.Extra))
		}

		exists, err := model.MessageExistsBySignature(title, fullContent, publishedAt, model.MessageSourceAdmin)
		if err != nil {
			common.SysError(fmt.Sprintf("failed to check legacy announcement duplication: %v", err))
			continue
		}
		if exists {
			continue
		}

		message := &model.Message{
			Title:       title,
			Content:     fullContent,
			Status:      model.MessageStatusOnline,
			Source:      model.MessageSourceAdmin,
			PublishedAt: &publishedAt,
		}
		if err = model.CreateMessage(message); err != nil {
			common.SysError(fmt.Sprintf("failed to migrate legacy announcement: %v", err))
			continue
		}
		if err = model.CreateUserMessageDeliveries(message.Id, recipients); err != nil {
			common.SysError(fmt.Sprintf("failed to create legacy announcement deliveries: %v", err))
			continue
		}
		migrated = true
	}
	return migrated
}

func migrateLegacyNoticeOption() bool {
	common.OptionMapRWMutex.RLock()
	raw := strings.TrimSpace(common.OptionMap["Notice"])
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return false
	}

	publishedAt := time.Now()
	title := deriveLegacyMessageTitle(raw, "", publishedAt)
	exists, err := model.MessageExistsBySignature(title, raw, publishedAt, model.MessageSourceAdmin)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to check legacy notice duplication: %v", err))
		return false
	}
	if exists {
		return false
	}

	recipients, err := model.ListEnabledUserRecipients()
	if err != nil {
		common.SysError(fmt.Sprintf("failed to list users for legacy notice migration: %v", err))
		return false
	}

	message := &model.Message{
		Title:       title,
		Content:     raw,
		Status:      model.MessageStatusOnline,
		Source:      model.MessageSourceAdmin,
		PublishedAt: &publishedAt,
	}
	if err = model.CreateMessage(message); err != nil {
		common.SysError(fmt.Sprintf("failed to migrate legacy notice: %v", err))
		return false
	}
	if err = model.CreateUserMessageDeliveries(message.Id, recipients); err != nil {
		common.SysError(fmt.Sprintf("failed to create legacy notice deliveries: %v", err))
		return false
	}
	return true
}

func deriveLegacyMessageTitle(content string, extra string, publishedAt time.Time) string {
	firstLine := strings.TrimSpace(strings.Split(content, "\n")[0])
	firstLine = strings.TrimLeft(firstLine, "#>-* ")
	if firstLine != "" {
		runes := []rune(firstLine)
		if len(runes) > 40 {
			firstLine = string(runes[:40])
		}
		return firstLine
	}
	if strings.TrimSpace(extra) != "" {
		return strings.TrimSpace(extra)
	}
	return fmt.Sprintf("系统消息 %s", publishedAt.Format("2006-01-02"))
}

func validateAdminMessageDraftInput(input *AdminMessageDraftInput) error {
	if input == nil {
		return errors.New("消息内容不能为空")
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" {
		return errors.New("标题和内容不能为空")
	}

	targetType := normalizeMessageTargetType(input.TargetType)
	switch targetType {
	case model.MessageTargetAll:
		input.TargetGroups = nil
		input.TargetUserIDs = nil
	case model.MessageTargetGroups:
		input.TargetGroups = normalizeMessageGroups(input.TargetGroups)
		input.TargetUserIDs = nil
		if len(input.TargetGroups) == 0 {
			return errors.New("请选择至少一个分组")
		}
	case model.MessageTargetUsers:
		input.TargetUserIDs = normalizeMessageUserIDs(input.TargetUserIDs)
		input.TargetGroups = nil
		if len(input.TargetUserIDs) == 0 {
			return errors.New("请选择至少一个用户")
		}
	default:
		return errors.New("无效的消息发送范围")
	}
	input.TargetType = targetType
	return nil
}

func normalizeMessageTargetType(targetType string) string {
	switch strings.TrimSpace(targetType) {
	case "", model.MessageTargetAll:
		return model.MessageTargetAll
	case model.MessageTargetGroups:
		return model.MessageTargetGroups
	case model.MessageTargetUsers:
		return model.MessageTargetUsers
	default:
		return strings.TrimSpace(targetType)
	}
}

func normalizeMessageGroups(groups []string) []string {
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

func normalizeMessageUserIDs(userIDs []int) []int {
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
