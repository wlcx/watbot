package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/koding/multiconfig"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)

type config struct {
	MumbleURL string `required:"true"`
	TgToken   string `required:"true"`
	TgChatID  int    `required:"true"`
	Debug     bool   `default:"false"`
}

func main() {
	conf := new(config)
	m := multiconfig.NewWithPath("config.toml")
	m.MustLoad(conf)

	bot, err := tgbotapi.NewBotAPI(conf.TgToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = conf.Debug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	if err = bot.UpdatesChan(u); err != nil {
		log.Panic(err)
	}

	// Mumble config
	mumbleParsedURL, err := url.Parse(conf.MumbleURL)
	if err != nil {
		log.Fatal(err)
	}
	if url.User == nil {
		log.Fatal("Mumble URL must include a username")
	}
	mumbleConf := gumble.NewConfig()
	if !strings.ContainsRune(mumbleParsedURL.Host, ':') { // If address does not specify port...
		mumbleConf.Address = mumbleParsedURL.Host + ":64738"
	} else {
		mumbleConf.Address = mumbleParsedURL.Host
	}
	mumbleConf.Username = mumbleParsedURL.User.Username()
	if pass, ok := mumbleParsedURL.User.Password(); ok {
		mumbleConf.Password = pass
	}
	client := gumble.NewClient(mumbleConf)

	client.Attach(gumbleutil.Listener{
		UserChange: func(e *gumble.UserChangeEvent) {
			if e.Type.Has(gumble.UserChangeConnected) {
				fmt.Println("User connected!")
				msgconf := tgbotapi.NewMessage(conf.TgChatID, fmt.Sprintf("%s connected #cgsnotify", e.User.Name))
				bot.SendMessage(msgconf)
			}
		},
	})
	if err := client.Connect(); err != nil {
		panic(err)
	}
	for update := range bot.Updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}
