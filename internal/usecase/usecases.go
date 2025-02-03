package usecase

import (
	"binance_test/internal/config"
	"binance_test/internal/ws_server"
)

type Usecases struct {
	Binance *BinanceUsecase
}

func CreateUsecases(conf *config.AppConfig, wsServer *ws_server.Server) *Usecases {
	return &Usecases{
		Binance: NewBinanceUsecase(conf, wsServer),
	}
}
