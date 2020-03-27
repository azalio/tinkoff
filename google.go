package main

import (
	"fmt"
	"github.com/TinkoffCreditSystems/invest-openapi-go-sdk"
	// "github.com/piquette/finance-go/quote"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// spreadsheetID from google docs
const spreadsheetID string = "1aEBxrtx21LvsONusIUeghL0cf_D5UjGLTabrFMJ_L5s"

func authInGoogle() *http.Client {
	data, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		panic(err.Error())
	}

	conf, err := google.JWTConfigFromJSON(data, spreadsheet.Scope)
	if err != nil {
		panic(err.Error())
	}

	client := conf.Client(context.TODO())
	return client
}

func getSpreadsheet(sheetName string) *spreadsheet.Sheet {
	client := authInGoogle()
	service := spreadsheet.NewServiceWithClient(client)

	spreadsheet, err := service.FetchSpreadsheet(spreadsheetID)
	if err != nil {
		panic(err.Error())
	}
	sheet, err := spreadsheet.SheetByTitle(sheetName)
	if err != nil {
		panic(err.Error())
	}
	return sheet
}

func updateGoogleSheet(account sdk.AccountType, portfolio sdk.Portfolio) {
	accountToString := fmt.Sprintf("%v", account)
	sheet := getSpreadsheet(accountToString)

	sheet.Update(0, 0, "Валюта")
	sheet.Update(0, 1, "Balance")
	sheet.Update(0, 2, "Blocked")

	// create Currencies table
	for idx, currencie := range portfolio.Currencies {
		sheet.Update(idx+1, 0, fmt.Sprintf("%v", currencie.Currency))
		sheet.Update(idx+1, 1, fmt.Sprintf("%v", currencie.Balance))
		sheet.Update(idx+1, 2, fmt.Sprintf("%v", currencie.Blocked))
		// log.Printf("%+v\n", currencie)
	}
	err := sheet.Synchronize()
	if err != nil {
		panic(err.Error())
	}

	// Ticker
	// Name
	// InstrumentType
	// Stock Balance
	// Blocked
	// Lots
	// ExpectedYield
	// ExpectedYield
	// Currency
	// AveragePositionPrice

	sharesTableTitles := []string{"FIGI", "Ticker", "ISIN", "InstrumentType", "Balance",
		"Blocked", "Lots", "ExpectedYield", "Currency", "AveragePositionPrice", "AveragePositionPriceNoNkd", "Name"}

	for idx, field := range sharesTableTitles {
		sheet.Update(0, idx+4, field)
	}

	totalTable := make(map[string]float64)
	usdPrice := getRegularMarketPrice("RUB=X")

	for idx, position := range portfolio.Positions {
		// update stock table
		sheet.Update(1+idx, 4, position.FIGI)
		sheet.Update(1+idx, 5, position.Ticker)
		sheet.Update(1+idx, 6, position.ISIN)
		sheet.Update(1+idx, 7, fmt.Sprintf("%v", position.InstrumentType))
		sheet.Update(1+idx, 8, fmt.Sprintf("%f", position.Balance))
		sheet.Update(1+idx, 9, fmt.Sprintf("%f", position.Blocked))
		sheet.Update(1+idx, 10, fmt.Sprintf("%d", position.Lots))
		sheet.Update(1+idx, 11, fmt.Sprintf("%f", position.ExpectedYield.Value))
		sheet.Update(1+idx, 12, fmt.Sprintf("%v", position.ExpectedYield.Currency))
		sheet.Update(1+idx, 13, fmt.Sprintf("%f", position.AveragePositionPrice.Value))
		sheet.Update(1+idx, 14, fmt.Sprintf("%f", position.AveragePositionPriceNoNkd.Value))
		sheet.Update(1+idx, 15, position.Name)

		// Update total tabl
		var positionPrice float64
		positionInstrumentType := fmt.Sprintf("%v", position.InstrumentType)
		if positionInstrumentType == "Stock" {
			if position.ExpectedYield.Currency == "RUB" {
				position.Ticker = position.Ticker + ".ME"
			}
			positionPrice = getRegularMarketPrice(position.Ticker)
			if account == "Tinkoff" && position.ExpectedYield.Currency == "RUB" {
				positionPrice = positionPrice / usdPrice
			}
		}
		if positionInstrumentType == "Currency" {
			positionPrice = position.Balance
			position.Balance = 1
		}
		if positionInstrumentType == "Bond" {
			positionPrice = position.ExpectedYield.Value/float64(position.Balance) + position.AveragePositionPrice.Value
		}
		// log.Printf("Symbol: %s, type: %v, price: %f", position.Ticker, position.InstrumentType, positionPrice)
		totalTable[fmt.Sprintf("%v", position.InstrumentType)] += positionPrice * float64(position.Balance)
	}

	sharesTableTitles = []string{"InstrumentType", "Value", "Percent"}

	for idx, field := range sharesTableTitles {
		sheet.Update(7, idx, field)
	}
	totalValue := totalTable["Currency"] + totalTable["Bond"] + totalTable["Stock"]
	sheet.Update(8, 0, "Stock")
	sheet.Update(8, 1, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Stock"]), ".", ","))
	sheet.Update(8, 2, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Stock"]/totalValue), ".", ","))
	sheet.Update(9, 0, "Bond")
	sheet.Update(9, 1, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Bond"]), ".", ","))
	sheet.Update(9, 2, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Bond"]/totalValue), ".", ","))
	sheet.Update(10, 0, "Currency")
	sheet.Update(10, 1, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Currency"]), ".", ","))
	sheet.Update(10, 2, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Currency"]/totalValue), ".", ","))
	sheet.Update(11, 0, "Total")
	sheet.Update(11, 1, strings.ReplaceAll(fmt.Sprintf("%f", totalTable["Currency"]+totalTable["Bond"]+totalTable["Stock"]), ".", ","))

	err = sheet.Synchronize()
	if err != nil {
		panic(err.Error())
	}
	// log.Println(totalTable)
}

func Chase(portfolioList []sdk.PositionBalance) {
	// Get data from the Chase table in google docs
	sheetName := "Chase"
	sheet := getSpreadsheet(sheetName)
	// log.Printf("Sheet rows: %d", len(sheet.Rows))
	// numOfRows := len(sheet.Rows)
	var symbols []string
	for idx, row := range sheet.Rows {
		// log.Printf("%v %v", idx, row)
		if idx == 0 {
			continue
		}
		symbol := row[0].Value
		// log.Printf("%s", symbol)
		if row[1].Value == "RUB" {
			symbol = symbol + ".ME" // for Yahoo
		}
		symbols = append(symbols, symbol)
	}
	// Get data about stocks
	idx := 1
	qIter := GetData(symbols)
	for qIter.Next() {
		q := qIter.Quote()
		// log.Printf("%v", q)
		if q.Symbol == sheet.Rows[idx][0].Value || q.Symbol == sheet.Rows[idx][0].Value+".ME" {
			// log.Printf("%s %f", q.Symbol, q.RegularMarketPrice)
			decision := buyOrSellMA(q)
			sheet.Update(idx, 2, fmt.Sprintf("%f", q.RegularMarketPrice))
			wantPrice, err := strconv.ParseFloat(sheet.Rows[idx][3].Value, 32)
			if err == nil {
				if wantPrice >= q.RegularMarketPrice {
					text := fmt.Sprintf("%s want price: %f, current price: %f", q.Symbol, wantPrice, q.RegularMarketPrice)
					log.Println(text)
					text = text + fmt.Sprintf("\nhttps://www.tradingview.com/symbols/%s", sheet.Rows[idx][0].Value)
					sendMessageICQ(text, icqto)
				}
			}
			sheet.Update(idx, 4, fmt.Sprintf("%f", q.FiftyDayAverage))
			if sheet.Rows[idx][5].Value != fmt.Sprintf("%t", decision.FiftyDayAverage) {
				text := fmt.Sprintf("%s FiftyDayAverage is changed. Was: %s, became: %t",
					sheet.Rows[idx][0].Value, sheet.Rows[idx][5].Value, decision.FiftyDayAverage)
				log.Println(text)
				text = text + fmt.Sprintf("\nhttps://www.tradingview.com/symbols/%s", sheet.Rows[idx][0].Value)
				sendMessageICQ(text, icqto)
			}
			sheet.Update(idx, 5, fmt.Sprintf("%t", decision.FiftyDayAverage))
			sheet.Update(idx, 6, fmt.Sprintf("%f", q.TwoHundredDayAverage))
			if sheet.Rows[idx][7].Value != fmt.Sprintf("%t", decision.TwoHundredDayAverage) {
				text := fmt.Sprintf("%s TwoHundredDayAverage is changed. Was: %s, became: %t",
					sheet.Rows[idx][0].Value, sheet.Rows[idx][5].Value, decision.TwoHundredDayAverage)
				log.Println(text)
				text = text + fmt.Sprintf("\nhttps://www.tradingview.com/symbols/%s", sheet.Rows[idx][0].Value)
				sendMessageICQ(text, icqto)
			}
			sheet.Update(idx, 7, fmt.Sprintf("%t", decision.TwoHundredDayAverage))
			sheet.Update(idx, 8, "https://www.tradingview.com/symbols/"+sheet.Rows[idx][0].Value)
			idx += 1
		}
	}

	var stocks []string
	for _, position := range portfolioList {
		// log.Printf("%+v\n", position)
		if position.InstrumentType == "Stock" {
			positionTickerRus := position.Ticker
			if position.AveragePositionPrice.Currency == "RUB" {
				// Yahoo finance wants YNDX.ME for MOEX
				positionTickerRus = position.Ticker + ".ME"
			}
			if !stringInSlice(positionTickerRus, symbols) {
				if !stringInSlice(position.Ticker, stocks) {
					stocks = append(stocks, position.Ticker)
				}
			}
		}
	}

	// log.Println("Symbols")
	// for _, symbol := range symbols {
	// 	log.Println(symbol)
	// }

	// log.Println("Stocks")
	// for _, stock := range stocks {
	// 	log.Println(stock)
	// }

	for idxStocks, stock := range stocks {
		sheet.Update(idx+idxStocks, 0, stock)
		// log.Printf("%s %v", stock, portfolioList[idxStocks].AveragePositionPrice.Currency)
		sheet.Update(idx+idxStocks, 1, fmt.Sprintf("%v", portfolioList[idxStocks].AveragePositionPrice.Currency))
	}

	err := sheet.Synchronize()
	if err != nil {
		panic(err.Error())
	}
}
