package telegram

import (
	"aroma-hub/internal/application/dto"
	"aroma-hub/internal/models"
	"context"
	"errors"
	"fmt"
	"log"
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
	adminIDs map[int64]struct{}
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
		ParseMode: telebot.ModeMarkdown,
		OnError: func(err error, c telebot.Context) {
			log.Printf("[Telebot Error] %v", err)
		},
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
		adminIDs: make(map[int64]struct{}),
	}, nil
}

func (p *TelegramProvider) registerCommands() {
}

func (p *TelegramProvider) Start() {
	p.EnrichAdmins(context.Background())

	p.registerCommands()

	p.bot.Start()
}

func (p *TelegramProvider) helloHandler(c telebot.Context) error {
	sender := c.Sender()
	greeting := fmt.Sprintf("Привіт, %s! 👋 Мені приємно з тобою поспілкуватися.", sender.FirstName)
	return c.Send(greeting)
}

func (p *TelegramProvider) Stop() {
	p.bot.Stop()
}

func (p *TelegramProvider) EnrichAdmins(ctx context.Context) error {
	admins, err := p.storage.ListAdmins(ctx, dto.ListAdminFilter{})
	if err != nil {
		return err
	}

	for _, admin := range admins {
		int64AdminVendorID, err := strconv.ParseInt(admin.VendorID, 10, 64)
		if err != nil {
			return err
		}

		log.Printf("Admin ID: %d", int64AdminVendorID)

		p.adminIDs[int64AdminVendorID] = struct{}{}
	}

	return nil
}

func (p *TelegramProvider) isAdmin(userID int64) bool {
	_, ok := p.adminIDs[userID]

	return ok
}
