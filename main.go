package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/layeh/gumble/gumble"
	"github.com/layeh/gumble/gumbleutil"
)

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
	godotenv.Load()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	// Mumble config
	mumbleParsedURL, err := url.Parse(os.Getenv("MUMBLE_URL"))
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
				chatid, err := strconv.Atoi(os.Getenv("TG_CHAT_ID"))
				if err != nil {
					log.Fatal(err)
				}
				msgconf := tgbotapi.NewMessage(chatid, fmt.Sprintf("%s connected #cgsnotify", e.User.Name))
				bot.Send(msgconf)
			}
		},
	})
	if err := client.Connect(); err != nil {
		panic(err)
	}
	for update := range updates {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		switch update.Message.Text {
		case "/users":
			var users []string
			for _, user := range client.Users {
				if user.Session != client.Self.Session { // Skip adding ourself to the online users
					users = append(users, user.Name)
				}
			}
			if len(users) == 0 {
				users = []string{"Noone"}
			}

			var isare string
			if len(users) < 2 {
				isare = "is"
			} else {
				isare = "are"
			}

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%s %s online", prettyListify(users), isare)))

		case "/chatid":
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("This chat's ID: %d", update.Message.Chat.ID)))
		}
	}
}
