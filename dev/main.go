package main

import (
	"log"
	"time"

	dc "github.com/bwmarrin/discordgo"
	"github.com/pure-nomad/cordkit"
)

func main() {
	bot, err := cordkit.NewBot("../client.json")
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

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

	conn := bot.HandleConnection("fakeconn1")
	time.Sleep(time.Second * 1)
	bot.KillConnection(conn)

	select {}
}
