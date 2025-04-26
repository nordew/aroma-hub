package telegram

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/telebot.v4"
)

const (
	defaultAccountName  = "admin"
	otpPrefix           = "otp_"
	otpTTL              = 5 * time.Minute
	accessDeniedMessage = "⛔ *Access denied*. You don't have permission to use this command."
)

func (p *TelegramProvider) RegisterLoginCommand() {
	p.bot.Handle("/login", func(c telebot.Context) error {
		user := c.Sender()

		userIDToString := fmt.Sprintf("%d", user.ID)

		isAdmin, err := p.service.IsAdmin(context.Background(), userIDToString)
		if err != nil {
			return err
		}
		if !isAdmin {
			return p.sendAccessDenied(user)
		}

		otpCode, err := p.otpGen.GenerateOTP(defaultAccountName)
		if err != nil {
			return err
		}

		go func() {
			p.cache.SetWithTTL(otpCode, userIDToString, otpTTL)
		}()

		msg := fmt.Sprintf("🔐 *One-Time Password*\n\n"+
			"Your OTP code: `%s`\n\n"+
			"⏱ Valid for *5 minutes*\n\n"+
			"_Use this code to log in to the admin panel._", otpCode)

		_, err = p.bot.Send(user, msg, &telebot.SendOptions{
			ParseMode: telebot.ModeMarkdown,
		})
		return err
	})
}

func (p *TelegramProvider) sendAccessDenied(user *telebot.User) error {
	_, err := p.bot.Send(user, accessDeniedMessage, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
	return err
}
