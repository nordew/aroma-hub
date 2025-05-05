package telegram

import (
	"aroma-hub/internal/application/dto"
	"context"
	"fmt"
	"strconv"
	"time"

	"gopkg.in/telebot.v4"
)

func (p *TelegramProvider) registerLoginCommand() {
	p.bot.Handle("/login", p.handleLogin)
}

func (p *TelegramProvider) handleLogin(c telebot.Context) error {
	ctx := context.Background()

	userID := c.Sender().ID
	vendorID := strconv.FormatInt(userID, 10)

	admins, err := p.storage.ListAdmins(ctx, dto.ListAdminFilter{VendorID: vendorID})
	if err != nil {
		return c.Send(fmt.Sprintf("Error checking admin status: %v", err))
	}

	if len(admins) == 0 {
		return c.Send("❌ You are not registered as an admin.")
	}

	otp, err := p.otpGen.GenerateOTP(vendorID)
	if err != nil {
		return c.Send(fmt.Sprintf("Error generating OTP: %v", err))
	}

	go p.cache.SetWithTTL(otp, vendorID, 5*time.Minute)

	msg := fmt.Sprintf("🔐 Your login OTP is: `%s`", otp)
	return c.Send(msg)
}
