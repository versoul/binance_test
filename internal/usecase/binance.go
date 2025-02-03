package usecase

import (
	"binance_test/internal/config"
	"binance_test/internal/entities"
	"binance_test/internal/ws_server"
	"encoding/json"
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/rs/zerolog/log"
	"math"
	"net/url"
	"strconv"
	"strings"
)

type BinanceUsecase struct {
	wsServer    *ws_server.Server
	conn        *websocket.Conn
	url         string
	symbolsData map[string]entities.SymbolsData
	symbols     []string
}

func NewBinanceUsecase(conf *config.AppConfig, wsServer *ws_server.Server) *BinanceUsecase {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:9443", Path: "/ws"}

	usecase := BinanceUsecase{
		wsServer: wsServer,
		url:      u.String(),
		symbols:  conf.Symbols,
	}

	usecase.symbolsData = make(map[string]entities.SymbolsData)
	for _, symbol := range conf.Symbols {
		data := entities.SymbolsData{}
		data.Bids = make(map[string]float64)
		data.Asks = make(map[string]float64)
		usecase.symbolsData[strings.ToLower(symbol)] = data
	}

	return &usecase
}

func (u *BinanceUsecase) Start() error {
	conn, _, err := websocket.DefaultDialer.Dial(u.url, nil)
	if err != nil {
		return err
	}
	u.conn = conn

	subscribeMessage := entities.BinanceRequest{
		Method: "SUBSCRIBE",
		ID:     1,
	}
	for _, symbol := range u.symbols {
		subscribeMessage.Params = append(subscribeMessage.Params,
			fmt.Sprintf("%s@depth", strings.ToLower(symbol)),
		)
	}
	err = conn.WriteJSON(subscribeMessage)
	if err != nil {
		return err
	}

	go func() {
		for {
			_, message, err := u.conn.ReadMessage()
			if err != nil {
				return
			}

			var data entities.BinanceAnswr
			if err := json.Unmarshal(message, &data); err != nil {
				log.Error().Err(err).Msgf("Unmarshal error")
				continue
			}
			u.ProcessData(data)
		}
	}()

	return nil
}

func (u *BinanceUsecase) Stop() {
	u.conn.Close()
}

func (u *BinanceUsecase) ProcessData(data entities.BinanceAnswr) {
	symbol := strings.ToLower(data.Symbol)
	symbolData, ok := u.symbolsData[symbol]
	if !ok {
		return
	}

	for _, bid := range data.Bids {
		priceStr := bid[0]
		quantity, _ := strconv.ParseFloat(bid[1], 64)
		if quantity == 0 {
			delete(symbolData.Bids, priceStr)
		} else {
			symbolData.Bids[priceStr] = quantity
		}
	}

	for _, ask := range data.Asks {
		priceStr := ask[0]
		quantity, _ := strconv.ParseFloat(ask[1], 64)
		if quantity == 0 {
			delete(symbolData.Asks, priceStr)
		} else {
			symbolData.Asks[priceStr] = quantity
		}
	}

	bestBid := 0.0
	for priceStr := range symbolData.Bids {
		price, _ := strconv.ParseFloat(priceStr, 64)
		if price > bestBid {
			bestBid = price
		}
	}

	bestAsk := math.MaxFloat64
	for priceStr := range symbolData.Asks {
		price, _ := strconv.ParseFloat(priceStr, 64)
		if price < bestAsk {
			bestAsk = price
		}
	}
	if bestAsk == math.MaxFloat64 {
		bestAsk = 0.0
	}

	if bestBid != symbolData.CurrentBid || bestAsk != symbolData.CurrentAsk {
		symbolData.CurrentBid = bestBid
		symbolData.CurrentAsk = bestAsk
		u.symbolsData[symbol] = symbolData
		u.SendDataUpdate(entities.Response{
			Symbol: strings.ToUpper(symbol),
			Ask:    symbolData.CurrentAsk,
			Bid:    symbolData.CurrentBid,
		})
	}
}

func (u *BinanceUsecase) SendDataInitial() {
	for symbol, data := range u.symbolsData {
		msg := entities.Response{
			Symbol: strings.ToUpper(symbol),
			Ask:    data.CurrentAsk,
			Bid:    data.CurrentBid,
		}
		u.SendDataUpdate(msg)
	}
}

func (u *BinanceUsecase) SendDataUpdate(msg entities.Response) {
	u.wsServer.WriteAll(msg)
}
