package telegram

import (
	"context"
	"errors"
	"time"

	stash "github.com/nordew/go-stash"
	"gopkg.in/telebot.v4"
)

type Service interface {
	IsAdmin(ctx context.Context, vendorID string) (bool, error)
}

type OTPGenerator interface {
	GenerateOTP(accountName string) (string, error)
}

var (
	ErrInvalidRecipientID = errors.New("invalid recipient ID")
	ErrSendingMessage     = errors.New("failed to send message")
)

type TelegramProvider struct {
	bot      *telebot.Bot
	apiToken string
	service  Service
	otpGen   OTPGenerator
	cache    stash.Cache
}

func NewTelegramProvider(
	apiToken string,
	service Service,
	otpGen OTPGenerator,
	cache stash.Cache,
) (*TelegramProvider, error) {
	pref := telebot.Settings{
		Token:     apiToken,
		Poller:    &telebot.LongPoller{Timeout: 10 * time.Second},
		ParseMode: telebot.ModeHTML,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return &TelegramProvider{
		bot:      bot,
		apiToken: apiToken,
		service:  service,
		otpGen:   otpGen,
		cache:    cache,
	}, nil
}

func (p *TelegramProvider) RegisterCommands() {
	p.RegisterLoginCommand()
}

func (p *TelegramProvider) Start() {
	go p.bot.Start()
}

func (p *TelegramProvider) Stop() {
	p.bot.Stop()
}
