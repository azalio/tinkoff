package main

import (
	"context"
	sdk "github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	"log"
	"time"
)

func rest() {
	client := sdk.NewRestClient(*token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// log.Println("Получение всех брокерских счетов")
	accounts, err := client.Accounts(ctx)
	if err != nil {
		log.Fatalln(errorHandle(err))
	}
	// log.Printf("%+v\n", accounts)

	var stocks []string

	for _, account := range accounts {
		// log.Printf("%s - %s\n", account.Type, account.ID)
		portfolioList := getPortfolio(client, account.ID)
		// log.Printf("%+v\n", portfolioList)
		updateGoogleSheet(account.Type, portfolioList)
		for _, position := range portfolioList.Positions {
			// log.Printf("%+v\n", position)
			if position.InstrumentType == "Stock" {
				if position.AveragePositionPrice.Currency == "RUB" {
					// Yahoo finance wants YNDX.ME for MOEX
					position.Ticker = position.Ticker + ".ME"
				}
				if !stringInSlice(position.Ticker, stocks) {
					stocks = append(stocks, position.Ticker)
				}
			}
		}
		// for _, currencie := range portfolioList.Currencies {
		// 	log.Printf("%+v\n", currencie)
		// }
	}

	// log.Println(stocks)
	iter := GetData(stocks)
	for iter.Next() {
		q := iter.Quote()
		decisions := buyOrSellMA(q)
		if err != nil {
			log.Panicf("Problem with buyOrSellMA function, err is :%s", err)
		}
		// } else {
		// 	log.Printf("Symbol is: %s, decision is: %v", q.Symbol, decisions)
		// }
		log.Printf("Symbol: %s, RegularMarketPrice: %f, FiftyDayAverage: %f, TwoHundredDayAverage: %f",
			q.Symbol, q.RegularMarketPrice, q.FiftyDayAverage, q.TwoHundredDayAverage)
		log.Printf("buy FiftyDayAverage: %t, buy TwoHundredDayAverage: %t",
			decisions.FiftyDayAverage, decisions.TwoHundredDayAverage)
	}

	// for _, share := range data {
	// 	log.Printf("%v", share)
	// }

}

func getPortfolio(client *sdk.RestClient, portfolio_id string) sdk.Portfolio {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// log.Println("Получение списка валютных и НЕ валютных активов портфеля для счета по-умолчанию")
	// Метод является совмещеним PositionsPortfolio и CurrenciesPortfolio
	portfolio, err := client.Portfolio(ctx, portfolio_id)
	if err != nil {
		log.Fatalln(err)
	}
	return portfolio
}

func errorHandle(err error) error {
	if err == nil {
		return nil
	}

	if tradingErr, ok := err.(sdk.TradingError); ok {
		if tradingErr.InvalidTokenSpace() {
			tradingErr.Hint = "Do you use sandbox token in production environment or vise verse?"
			return tradingErr
		}
	}

	return err
}
