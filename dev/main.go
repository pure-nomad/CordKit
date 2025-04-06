package main

import (
	"log"
	"time"

	"github.com/pure-nomad/cordkit"
)

func main() {
	cordkit.HelloCordkit()
	clientSettings := cordkit.NewClient("xxx", "1358227184719757312", "1358230105838850239", "1358230135295180910", "active", "dead")
	bot := cordkit.NewBot(clientSettings)
	bot.Running = true
	defer bot.Stop()
	bot.Start()
	log.Println("Bot is online")

	client := bot.Client

	if bot.Running {
		conn := client.HandleConnection("fakeconn1")
		time.Sleep(time.Second * 1)
		if bot.Running {
			client.KillConnection(conn)
		}
	}

	select {}
}
