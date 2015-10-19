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

// prettyListify takes a slice of strings and returns a string of the form "x, y and z"
func prettyListify(things []string) string {
	if len(things) <= 2 {
		return strings.Join(things, " and ")
	} else {
		thing, things := things[0], things[1:]
		return thing + ", " + prettyListify(things)
	}
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
		switch update.Message.Text {
		case "/users":
			var users []string
			for userid, user := range client.Users {
				if userid != client.Self.UserID { // Skip adding ourself to the online users
					users = append(users, user.Name)
				}
			}
			bot.SendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, prettyListify(users)+" are online"))
		case "/chatid":
			bot.SendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("This chat's ID: %d", update.Message.Chat.ID)))
		}
	}
}
