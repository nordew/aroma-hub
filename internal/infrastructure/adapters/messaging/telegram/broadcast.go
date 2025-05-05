package telegram

import (
	"context"

	"gopkg.in/telebot.v4"
)

func (p *TelegramProvider) BroadcastMessage(ctx context.Context, text string) error {
	if len(p.adminIDs) == 0 {
		return ErrInvalidRecipientID
	}

	for id := range p.adminIDs {
		if _, err := p.bot.Send(telebot.ChatID(id), text); err != nil {
			return ErrSendingMessage
		}
	}

	return nil
}
