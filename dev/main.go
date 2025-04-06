package main

import (
	"log"
	"time"

	dc "github.com/bwmarrin/discordgo"
	"github.com/pure-nomad/cordkit"
)

func main() {
	cordkit.HelloCordkit()
	clientSettings := cordkit.NewClient("xxx", "1358227184719757312", "1358230105838850239", "1358230135295180910", "1358501105906090117", "active", "dead")
	bot := cordkit.NewBot(clientSettings, true)
	bot.CustomCommands = true

	bot.Commands = append(bot.Commands, cordkit.Command{
		Name:        "ping",
		Description: "Responds with pong",
		Action: func(b *cordkit.Bot, i *dc.InteractionCreate) {
			b.SendMsg(i.ChannelID, "Pong!")
		},
	})

	defer bot.Stop()
	bot.Start()
	log.Println("Bot is online")

	if bot.Running {
		conn := bot.HandleConnection("fakeconn1")
		time.Sleep(time.Second * 1)
		if bot.Running {
			bot.KillConnection(conn)
		}
	}

	select {}
}
