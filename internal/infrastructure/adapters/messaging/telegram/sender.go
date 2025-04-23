package telegram

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/telebot.v4"
)

func (p *TelegramProvider) SendMessage(ctx context.Context, recipientID string, text string) (string, error) {
	chatID, err := p.parseRecipientID(recipientID)
	if err != nil {
		return "", err
	}

	recipient := &telebot.User{ID: chatID}

	var sentMsg *telebot.Message

	done := make(chan struct{})
	var sendErr error

	go func() {
		defer close(done)
		sentMsg, sendErr = p.bot.Send(recipient, text)
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("send operation cancelled: %w", ctx.Err())
	case <-done:
		if sendErr != nil {
			return "", fmt.Errorf("%w: %v", ErrSendingMessage, sendErr)
		}
		return strconv.Itoa(sentMsg.ID), nil
	}
}

func (p *TelegramProvider) parseRecipientID(id string) (int64, error) {
	chatID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidRecipientID, id)
	}
	return chatID, nil
}
