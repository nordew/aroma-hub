package telegram

import "aroma-hub/internal/infrastructure/adapters/external_api"

type telegramAPI struct {
}

func NewTelegramAPI() external_api.ExternalAPI {
	return &telegramAPI{}
}
