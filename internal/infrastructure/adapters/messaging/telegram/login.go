package telegram

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/telebot.v4"
)

func (p *TelegramProvider) registerLoginCommand() {
	p.bot.Handle("/login", p.handleLogin)
}

func (p *TelegramProvider) handleLogin(c telebot.Context) error {
	userID := c.Sender().ID
	vendorID := strconv.FormatInt(userID, 10)

	if ok := p.isAdmin(userID); !ok {
		return c.Send("You are not authorized to login.")
	}

	otp, err := p.otpGen.GenerateOTP(vendorID)
	if err != nil {
		return c.Send(fmt.Sprintf("Error generating OTP: %v", err))
	}

	go p.cache.SetWithTTL(otp, vendorID, 5*time.Minute)

	msg := fmt.Sprintf("🔐 Your login OTP is: `%s`", otp)
	return c.Send(msg)
}
