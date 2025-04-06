package cordkit

import (
	"fmt"
	"log"
	"os"
	"time"

	dc "github.com/bwmarrin/discordgo"
)

type Bot struct {
	Client         *Client
	Running        bool
	Logging        bool
	LogChannelID   string
	CustomCommands bool
	Commands       []Command
}

type Command struct {
	Name        string
	Description string
	Action      func(*Bot, *dc.InteractionCreate)
}

func NewBot(clientSettings *Client, logs bool) *Bot {
	return &Bot{
		Client:         clientSettings,
		Running:        true,
		Logging:        logs,
		CustomCommands: false,
		Commands:       []Command{},
	}
}

func (b *Bot) Start() {
	sess, err := dc.New("Bot " + b.Client.botToken)
	if err != nil {
		panic(err)
	}

	b.Client.botRef = sess
	b.Client.botRef.AddHandler(b.handleSlash)

	err = b.Client.botRef.Open()
	if err != nil {
		panic(err)
	}

	if b.Logging {

		transcriptName := fmt.Sprintf("transcript-%05d", time.Now().UnixNano()%100000)
		logChannel, err := b.Client.botRef.GuildChannelCreateComplex(b.Client.guildID, dc.GuildChannelCreateData{
			Name:     transcriptName,
			ParentID: b.Client.transcriptCategoryID,
		})

		if err != nil {
			panic(err)
		}

		b.LogChannelID = logChannel.ID
		log.Println("Created logging channel with ID: ", logChannel.ID)

	}

	now := time.Now()
	botStartMSG := fmt.Sprintf("Bot Started at %v", now.Format("03:04PM"))
	b.SendInfoLog(botStartMSG)

	cmds := []*dc.ApplicationCommand{
		{Name: "start", Description: "Enable bot logic"},
		{Name: "stop", Description: "Disable bot logic"},
		{Name: "purge", Description: "Delete all channels in the dead category"},
		{Name: "nuke", Description: "Delete all channels and stop the bot"},
	}

	if b.CustomCommands {
		for _, cmd := range b.Commands {
			cmds = append(cmds, &dc.ApplicationCommand{
				Name:        cmd.Name,
				Description: cmd.Description,
			})
		}
	}

	_, err = b.Client.botRef.ApplicationCommandBulkOverwrite(sess.State.User.ID, b.Client.guildID, cmds)
	if err != nil {
		panic(err)
	}
}

func (b *Bot) Stop() error {
	b.Running = false
	now := time.Now()
	botEndMSG := fmt.Sprintf("Bot stopped at %v", now.Format("03:04PM"))
	if b.Logging {
		b.SendErrorLog(botEndMSG)
	}
	return b.Client.botRef.Close()
}

func (b *Bot) SendInfoLog(content string) (*dc.Message, error) {
	msgContent := fmt.Sprintf("[✅] %s", content)
	return b.SendMsg(b.LogChannelID, msgContent)
}

func (b *Bot) SendErrorLog(content string) (*dc.Message, error) {
	msgContent := fmt.Sprintf("[❌] %s", content)
	return b.SendMsg(b.LogChannelID, msgContent)
}

func (b *Bot) handleSlash(s *dc.Session, i *dc.InteractionCreate) {
	if i.Type != dc.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "start":
		if b.Running {
			b.BotRespond(i, "Already running.")
			return
		}
		b.Running = true
		b.BotRespond(i, "Started.")

	case "stop":
		if !b.Running {
			b.BotRespond(i, "Already stopped.")
			return
		}
		b.Running = false
		b.BotRespond(i, "Stopped.")

	case "purge":
		channels, err := s.GuildChannels(b.Client.guildID)
		if err != nil {
			b.BotRespond(i, "Error fetching channels: "+err.Error())
			return
		}

		deletedCount := 0
		for _, channel := range channels {
			if channel.ParentID == b.Client.deadCategoryID {
				_, err := b.DeleteChannel(channel.ID)
				if err != nil {
					b.BotRespond(i, fmt.Sprintf("Error deleting channel %s: %s", channel.Name, err.Error()))
					return
				}
				deletedCount++
			}
		}

		msg := fmt.Sprintf("Purged %d channels", deletedCount)
		b.BotRespond(i, msg)
		if b.Logging {
			b.SendInfoLog(msg)
		}

	case "nuke":
		channels, err := s.GuildChannels(b.Client.guildID)
		if err != nil {
			b.BotRespond(i, "Error fetching channels: "+err.Error())
			return
		}

		deletedCount := 0
		for _, channel := range channels {
			if channel.Type == dc.ChannelTypeGuildCategory {
				continue
			}

			_, err := b.DeleteChannel(channel.ID)
			if err != nil {
				b.BotRespond(i, fmt.Sprintf("Error deleting channel %s: %s", channel.Name, err.Error()))
				return
			}
			deletedCount++
		}

		b.Running = false
		if err := b.Client.botRef.Close(); err != nil {
			b.BotRespond(i, "Error stopping bot: "+err.Error())
			return
		}

		os.Exit(0)

	default:
		if b.CustomCommands {
			for _, cmd := range b.Commands {
				if cmd.Name == i.ApplicationCommandData().Name {
					cmd.Action(b, i)
					return
				}
			}
		}
	}
}

func (b *Bot) BotRespond(i *dc.InteractionCreate, msg string) {
	b.Client.botRef.InteractionRespond(i.Interaction, &dc.InteractionResponse{
		Type: dc.InteractionResponseChannelMessageWithSource,
		Data: &dc.InteractionResponseData{Content: msg},
	})
}

type Connection struct {
	id        string
	channelID string
	createdAt time.Time
	lastSeen  time.Time
	status    string
	// Transcript []string
	// metadata map[string]string
}

func (b *Bot) HandleConnection(id string) *Connection {

	newConnChannel, err := b.CreateChannel(id)
	if err != nil {
		panic(err)
	}

	now := time.Now()
	newConnMSG := fmt.Sprintf("New Connection %s\nBegan at %v", id, now.Format("03:04PM"))

	if b.Logging {
		b.SendInfoLog(newConnMSG)
	}

	b.SendMsg(newConnChannel.ID, newConnMSG)

	return &Connection{
		id:        id,
		createdAt: time.Now(),
		status:    "active",
		channelID: newConnChannel.ID,
	}
}

func (b *Bot) KillConnection(conn *Connection) *Connection {
	deadChannel, err := b.MakeChannelDead(conn.channelID)
	if err != nil {
		panic(err)
	}

	now := time.Now()

	deadConnMsg := fmt.Sprintf("Connection %v died\nEnded at %v", conn.id, now.Format("03:04PM"))

	if b.Logging {
		b.SendErrorLog(deadConnMsg)
	}

	b.SendMsg(deadChannel.ID, deadConnMsg)

	conn = &Connection{
		lastSeen:  time.Now(),
		status:    "dead",
		channelID: deadChannel.ID,
	}

	return conn
}

type Client struct {
	botToken             string
	guildID              string
	activeCategoryID     string
	deadCategoryID       string
	transcriptCategoryID string
	botRef               *dc.Session
	activeChannelPrefix  string
	deadChannelPrefix    string
}

func NewClient(token string, guildID string, activeCategory string, deadCategory string, transcriptCategory string, activeChannelPrefix string, deadChannelPrefix string) *Client {

	c := Client{
		botToken:             token,
		guildID:              guildID,
		activeCategoryID:     activeCategory,
		deadCategoryID:       deadCategory,
		transcriptCategoryID: transcriptCategory,
		activeChannelPrefix:  activeChannelPrefix,
		deadChannelPrefix:    deadChannelPrefix,
	}

	return &c
}

func (b *Bot) SendMsg(channelID, content string) (*dc.Message, error) {
	return b.Client.botRef.ChannelMessageSend(channelID, content)
}

func (b *Bot) CreateChannel(name string) (*dc.Channel, error) {
	return b.Client.botRef.GuildChannelCreateComplex(b.Client.guildID, dc.GuildChannelCreateData{
		ParentID: b.Client.activeCategoryID,
		Name:     b.Client.activeChannelPrefix + "-" + name,
		Type:     dc.ChannelTypeGuildText,
	})
}

func (b *Bot) DeleteChannel(channelID string) (*dc.Channel, error) {
	return b.Client.botRef.ChannelDelete(channelID)
}

func (b *Bot) MakeChannelActive(channelID string) (*dc.Channel, error) {
	channel, err := b.Client.botRef.Channel(channelID)
	if err != nil {
		return nil, err
	}

	// Remove dead prefix if present and add active prefix
	name := channel.Name
	if len(b.Client.deadChannelPrefix) > 0 && len(name) > len(b.Client.deadChannelPrefix)+1 {
		name = name[len(b.Client.deadChannelPrefix)+1:]
	}

	return b.Client.botRef.ChannelEdit(channelID, &dc.ChannelEdit{
		ParentID: b.Client.activeCategoryID,
		Name:     b.Client.activeChannelPrefix + "-" + name,
	})
}

func (b *Bot) MakeChannelDead(channelID string) (*dc.Channel, error) {
	channel, err := b.Client.botRef.Channel(channelID)
	if err != nil {
		return nil, err
	}

	// Remove active prefix if present and add dead prefix
	name := channel.Name
	if len(b.Client.activeChannelPrefix) > 0 && len(name) > len(b.Client.activeChannelPrefix)+1 {
		name = name[len(b.Client.activeChannelPrefix)+1:]
	}

	return b.Client.botRef.ChannelEdit(channelID, &dc.ChannelEdit{
		ParentID: b.Client.deadCategoryID,
		Name:     b.Client.deadChannelPrefix + "-" + name,
	})
}
