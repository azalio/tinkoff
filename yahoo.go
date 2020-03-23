package main

import (
	"github.com/piquette/finance-go/quote"
	"github.com/piquette/finance-go"
)

type decision struct {
	symbol               string
	FiftyDayAverage      bool
	TwoHundredDayAverage bool
}

func GetData(symbols []string) *quote.Iter {
	iter := quote.List(symbols)
	return iter
}

func getRegularMarketPrice(symbol string) float64 {
	// log.Println(symbol)
	q, err := quote.Get(symbol)
	if err != nil {
		// Uh-oh!
		panic(err)
	}
	// log.Println(q)
	return q.RegularMarketPrice
}

// Use SMA50 SMA200 to decide
// return false if sell and true if buy
func buyOrSellMA(stockData *finance.Quote) decision {
	var dec decision
	dec.symbol = stockData.Symbol
	if stockData.FiftyDayAverage >= stockData.RegularMarketPrice {
		dec.FiftyDayAverage = false
	} else {
		dec.FiftyDayAverage = true
	}
	if stockData.TwoHundredDayAverage < stockData.RegularMarketPrice {
		dec.TwoHundredDayAverage = true
	} else {
		dec.TwoHundredDayAverage = false
	}
	return dec
}
