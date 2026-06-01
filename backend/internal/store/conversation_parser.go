package store

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

var senderTimePattern = regexp.MustCompile(`^(.+?)\s+(\d{1,2}:\d{2})\s+(.+)$`)

func parseUploadedConversation(uploadID string, content string, baseTime time.Time) []domain.ConversationMessage {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	messages := make([]domain.ConversationMessage, 0, len(lines))
	offset := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		senderName, body := parseConversationLine(line)
		if body == "" {
			continue
		}
		offset++
		messages = append(messages, domain.ConversationMessage{
			ID:         fmt.Sprintf("msg_%s_%03d", uploadID, offset),
			SenderType: inferSenderType(senderName),
			SenderName: senderName,
			Content:    body,
			SentAt:     formatTime(baseTime.Add(time.Duration(offset-1) * time.Second)),
		})
	}

	if len(messages) == 0 {
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			return []domain.ConversationMessage{}
		}
		customerName := inferUploadedCustomerName(content)
		return []domain.ConversationMessage{
			{
				ID:         "msg_" + uploadID + "_001",
				SenderType: "customer",
				SenderName: customerName,
				Content:    trimmed,
				SentAt:     formatTime(baseTime),
			},
		}
	}

	return messages
}

func parseConversationLine(line string) (string, string) {
	if matches := senderTimePattern.FindStringSubmatch(line); len(matches) == 4 {
		return cleanSenderName(matches[1]), strings.TrimSpace(matches[3])
	}

	for _, sep := range []string{"：", ":"} {
		if before, after, ok := strings.Cut(line, sep); ok {
			after = strings.TrimSpace(after)
			if after != "" {
				return cleanSenderName(before), after
			}
		}
	}

	fields := strings.Fields(line)
	if len(fields) >= 2 {
		return cleanSenderName(fields[0]), strings.Join(fields[1:], " ")
	}
	return inferUploadedCustomerName(line), strings.TrimSpace(line)
}

func inferSenderType(senderName string) string {
	senderName = strings.TrimSpace(senderName)
	if strings.Contains(senderName, "销售") || strings.EqualFold(senderName, "sales") || strings.HasPrefix(strings.ToLower(senderName), "agent") {
		return "sales"
	}
	return "customer"
}

func inferUploadedCustomerName(content string) string {
	for _, message := range parseUploadedConversation("infer", content, time.Now().UTC()) {
		if message.SenderType == "customer" && strings.TrimSpace(message.SenderName) != "" {
			return message.SenderName
		}
	}

	content = strings.TrimSpace(content)
	fields := strings.Fields(content)
	if len(fields) > 0 {
		first := cleanSenderName(fields[0])
		if first != "" {
			return first
		}
	}
	return "新客户"
}

func cleanSenderName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "：:，,。 ")
	if value == "" {
		return "新客户"
	}
	return value
}

func parsedOwnerName(ownerID string) string {
	switch ownerID {
	case "usr_002":
		return "销售B"
	default:
		return "销售A"
	}
}
