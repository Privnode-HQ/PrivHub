package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

type messageRequest struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	TargetType    string   `json:"target_type"`
	TargetGroups  []string `json:"target_groups"`
	TargetUserIDs []int    `json:"target_user_ids"`
}

func GetMessageTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"template": common.GetMessageTemplateHTML(),
			"placeholders": []string{
				"{{.Title}}",
				"{{.SystemName}}",
				"{{.RecipientName}}",
				"{{.RecipientEmail}}",
				"{{.ContentHTML}}",
				"{{.DashboardURL}}",
				"{{.NoticeText}}",
				"{{.CurrentYear}}",
				"{{.TermsURL}}",
				"{{.PrivacyURL}}",
			},
		},
	})
}

func UpdateMessageTemplate(c *gin.Context) {
	var req struct {
		Template string `json:"template"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的参数"})
		return
	}
	if err := common.ValidateMessageTemplate(req.Template); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.UpdateOption(common.MessageTemplateOptionKey, req.Template); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "模板已更新"})
}

func GetAdminMessages(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	messages, total, err := model.GetAdminMessages(pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	messageIDs := make([]uint, 0, len(messages))
	for _, message := range messages {
		messageIDs = append(messageIDs, message.Id)
	}
	stats, err := model.GetMessageDeliveryStats(messageIDs)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	username := c.GetString("username")
	items := make([]gin.H, 0, len(messages))
	for _, message := range messages {
		renderedHTML, renderErr := common.RenderMessageHTML(message.Title, message.Content, common.MessageTemplateContext{
			RecipientName: username,
			PublishedAt:   messageTime(message),
		})
		if renderErr != nil {
			renderedHTML = ""
		}

		messageStats := stats[message.Id]
		targetUserOptions, targetUserErr := model.GetMessageTargetUserOptions(message.GetTargetUserIDs())
		if targetUserErr != nil {
			targetUserOptions = []model.MessageTargetUserOption{}
		}
		items = append(items, gin.H{
			"id":                  message.Id,
			"title":               message.Title,
			"content":             message.Content,
			"status":              message.Status,
			"source":              message.Source,
			"target_type":         message.GetTargetType(),
			"target_groups":       message.GetTargetGroups(),
			"target_user_ids":     message.GetTargetUserIDs(),
			"target_user_options": targetUserOptions,
			"published_at":        message.PublishedAt,
			"created_at":          message.CreatedAt,
			"updated_at":          message.UpdatedAt,
			"html_content":        renderedHTML,
			"delivery_total":      messageStats["total"],
			"read_total":          messageStats["read_total"],
			"email_sent":          messageStats["email_sent"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": items,
			"total": total,
		},
	})
}

func CreateMessage(c *gin.Context) {
	var req messageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的参数"})
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "标题和内容不能为空"})
		return
	}
	message, err := service.CreateAdminDraftMessage(service.AdminMessageDraftInput{
		Title:         req.Title,
		Content:       req.Content,
		TargetType:    req.TargetType,
		TargetGroups:  req.TargetGroups,
		TargetUserIDs: req.TargetUserIDs,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "草稿已创建", "data": message})
}

func UpdateMessage(c *gin.Context) {
	messageID, err := parseMessageID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的消息ID"})
		return
	}

	var req messageRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的参数"})
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "标题和内容不能为空"})
		return
	}

	message, err := service.UpdateAdminDraftMessage(messageID, service.AdminMessageDraftInput{
		Title:         req.Title,
		Content:       req.Content,
		TargetType:    req.TargetType,
		TargetGroups:  req.TargetGroups,
		TargetUserIDs: req.TargetUserIDs,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "草稿已更新", "data": message})
}

func PublishMessage(c *gin.Context) {
	messageID, err := parseMessageID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的消息ID"})
		return
	}
	message, err := service.PublishAdminMessage(messageID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "消息已上线", "data": message})
}

func CopyMessage(c *gin.Context) {
	messageID, err := parseMessageID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的消息ID"})
		return
	}

	message, err := service.CopyAdminMessage(messageID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "草稿已复制", "data": message})
}

func DeleteMessage(c *gin.Context) {
	messageID, err := parseMessageID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的消息ID"})
		return
	}
	if err = model.DeleteDraftMessage(messageID); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "草稿已删除"})
}

func GetMyMessages(c *gin.Context) {
	userID := c.GetInt("id")
	username := c.GetString("username")
	userEmail, _ := model.GetUserEmail(userID)
	pageInfo := common.GetPageQuery(c)

	items, total, err := model.GetUserMessages(userID, pageInfo)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		renderedHTML, renderErr := common.RenderMessageHTML(item.Title, item.Content, common.MessageTemplateContext{
			RecipientName:  username,
			RecipientEmail: userEmail,
			PublishedAt:    userMessageTime(item),
		})
		if renderErr != nil {
			renderedHTML = ""
		}

		result = append(result, gin.H{
			"id":            item.Id,
			"title":         item.Title,
			"content":       item.Content,
			"status":        item.Status,
			"source":        item.Source,
			"published_at":  item.PublishedAt,
			"created_at":    item.CreatedAt,
			"read_at":       item.ReadAt,
			"email_sent_at": item.EmailSentAt,
			"html_content":  renderedHTML,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": result,
			"total": total,
		},
	})
}

func GetMyUnreadMessageCount(c *gin.Context) {
	userID := c.GetInt("id")
	total, err := model.CountUnreadUserMessages(userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "", "data": total})
}

func ReadMyMessage(c *gin.Context) {
	userID := c.GetInt("id")
	messageID, err := parseMessageID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的消息ID"})
		return
	}
	if err = model.MarkUserMessageRead(userID, messageID); err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "消息已标记为已读"})
}

func parseMessageID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func messageTime(message *model.Message) time.Time {
	if message.PublishedAt != nil {
		return *message.PublishedAt
	}
	return message.CreatedAt
}

func userMessageTime(message *model.UserMessageView) time.Time {
	if message.PublishedAt != nil {
		return *message.PublishedAt
	}
	return message.CreatedAt
}
