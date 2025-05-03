package telegram

import (
	"context"

	"gopkg.in/telebot.v4"
)

func (s *TelegramProvider) BroadcastMessage(ctx context.Context, text string) error {
	for _, adminID := range s.adminIDs {
		if _, err := s.bot.Send(telebot.ChatID(adminID), text); err != nil {
			return err
		}
	}

	return nil
}
