package telegram

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"errors"
	"strconv"
	"time"

	stash "github.com/nordew/go-stash"
	"gopkg.in/telebot.v4"
)

type Storage interface {
	ListAdmins(ctx context.Context, filter dto.ListAdminFilter) ([]models.Admin, error)
	ListOrders(ctx context.Context, filter dto.ListOrderFilter) ([]models.Order, int64, error)
	UpdateOrder(ctx context.Context, input dto.UpdateOrderRequest) error
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
	storage  Storage
	otpGen   OTPGenerator
	cache    stash.Cache
	adminIDs []int64
}

func NewTelegramProvider(
	apiToken string,
	storage Storage,
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
		storage:  storage,
		otpGen:   otpGen,
		cache:    cache,
		adminIDs: []int64{},
	}, nil
}

func (p *TelegramProvider) RegisterCommands() {
	// p.RegisterLoginCommand()
}

func (p *TelegramProvider) Start() {
	p.EnrichAdmins(context.Background())

	go p.bot.Start()
}

func (p *TelegramProvider) EnrichAdmins(ctx context.Context) error {
	admins, err := p.storage.ListAdmins(ctx, dto.ListAdminFilter{})
	if err != nil {
		return err
	}

	for _, admin := range admins {
		int64AdminID, err := strconv.ParseInt(admin.ID, 10, 64)
		if err != nil {
			return err
		}

		p.adminIDs = append(p.adminIDs, int64AdminID)
	}

	return nil
}

func (p *TelegramProvider) Stop() {
	p.bot.Stop()
}
