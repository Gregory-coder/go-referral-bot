package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)


func main() {
	configPath := "config.json"
	tablePath := "links.xlsx"

	conf, err := OpenConfig(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	t, err := NewTable(tablePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer t.Close()

	bot, err := telego.NewBot(conf.Token)
	if err != nil {
		log.Fatalln(err)
	}

	updates, _ := bot.UpdatesViaLongPolling(nil)

	bh, _ := th.NewBotHandler(bot, updates)

	defer bh.Stop()
	defer bot.StopLongPolling()

	// On /start
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		_, err := bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			fmt.Sprintf("Hi, %s!\nPress the button below, to get the link", update.Message.From.FirstName),
		).WithReplyMarkup(tu.InlineKeyboard(
			tu.InlineKeyboardRow(
				tu.InlineKeyboardButton("Create link").WithCallbackData("new_link"),
			)),
		))
		if err != nil {
			log.Println(err)
		}
	}, th.CommandEqual("start"))

	// On button pressed
	bh.HandleCallbackQuery(func(bot *telego.Bot, query telego.CallbackQuery) {
		link, err := bot.CreateChatInviteLink(&telego.CreateChatInviteLinkParams{
			ChatID: tu.ID(conf.Channel),
			Name: query.From.FirstName + " " + query.From.LastName,
		})

		if err == nil {
			_, err = bot.SendMessage(tu.Message(tu.ID(query.From.ID), link.InviteLink))
		}
		if err == nil {
			err = t.AddRecord(query.From.FirstName, query.From.LastName, query.From.Username, link.InviteLink)
		}
		if err == nil {
			err = bot.AnswerCallbackQuery(tu.CallbackQuery(query.ID).WithText("Done"))
		}

		if err != nil {
			bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Something went wrong"))
			log.Println(err)
		}
	}, th.AnyCallbackQueryWithMessage(), th.CallbackDataEqual("new_link"))
	
	// On /links
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		if !slices.Contains(conf.Admins, update.Message.Chat.ID) {
			bot.SendMessage(tu.Message(tu.ID(update.Message.Chat.ID), "This function is unavailable for you"))
			return
		}

		file, err := os.Open(tablePath)
		if err == nil {
			document := tu.Document(
				tu.ID(update.Message.Chat.ID),
				tu.File(file),
			).WithCaption("The list of generated links")
		
			_, err = bot.SendDocument(document)
		}
		if err != nil {
			bot.SendMessage(tu.Message(tu.ID(update.Message.Chat.ID), "Something went wrong"))
			log.Println(err)
		}
	}, th.CommandEqual("links"))

	// On any other commands 
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"Unknown command, use /start",
		))
	}, th.AnyCommand())

	// On any other messages
	bh.Handle(func(bot *telego.Bot, update telego.Update) {
		bot.SendMessage(tu.Message(
			tu.ID(update.Message.Chat.ID),
			"I don't understand you, use /start",
		))
	}, th.AnyMessage())

	log.Println("Bot successfully started")

	bh.Start()
}