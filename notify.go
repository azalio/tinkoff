package main

import (
	"github.com/mail-ru-im/bot-golang"
	"log"
)

var bot *botgolang.Bot

func authICQ(BOT_TOKEN string) {
	var err error
    bot, err = botgolang.NewBot(BOT_TOKEN)
    if err != nil {
        log.Println("wrong token")
    }
}

func sendMessageICQ(text string, to string) {
    message := bot.NewTextMessage(to, text)
    message.Send()
}