package cordkit

import (
	"fmt"
	"time"

	dc "github.com/bwmarrin/discordgo"
)

type Bot struct {
	Client  *Client
	Running bool
}

func NewBot(clientSettings *Client) *Bot {
	return &Bot{
		Client:  clientSettings,
		Running: false,
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

	cmds := []*dc.ApplicationCommand{
		{Name: "start", Description: "Enable bot logic"},
		{Name: "stop", Description: "Disable bot logic"},
	}
	_, err = b.Client.botRef.ApplicationCommandBulkOverwrite(sess.State.User.ID, b.Client.guildID, cmds)
	if err != nil {
		panic(err)
	}
}

func (b *Bot) Stop() error {
	b.Running = false
	return b.Client.botRef.Close()
}

func (b *Bot) handleSlash(s *dc.Session, i *dc.InteractionCreate) {
	if i.Type != dc.InteractionApplicationCommand {
		return
	}

	switch i.ApplicationCommandData().Name {
	case "start":
		if b.Running {
			respond(s, i, "Already running.")
			return
		}
		b.Running = true
		respond(s, i, "Started.")

	case "stop":
		if !b.Running {
			respond(s, i, "Already stopped.")
			return
		}
		b.Running = false
		respond(s, i, "Stopped.")
	}
}

func respond(s *dc.Session, i *dc.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &dc.InteractionResponse{
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

func (c *Client) HandleConnection(id string) *Connection {

	newConnChannel, err := c.CreateChannel(id)
	if err != nil {
		panic(err)
	}

	newConnMsg := fmt.Sprintf("New Connection %v", id)
	c.SendMsg(newConnChannel.ID, newConnMsg)

	return &Connection{
		id:        id,
		createdAt: time.Now(),
		status:    "active",
		channelID: newConnChannel.ID,
	}
}

func (c *Client) KillConnection(conn *Connection) *Connection {
	deadChannel, err := c.MakeChannelDead(conn.channelID)
	if err != nil {
		panic(err)
	}

	deadConnMsg := fmt.Sprintf("Connection %v died", conn.id)
	c.SendMsg(deadChannel.ID, deadConnMsg)

	conn = &Connection{
		lastSeen:  time.Now(),
		status:    "dead",
		channelID: deadChannel.ID,
	}

	return conn
}

type Client struct {
	botToken            string
	guildID             string
	activeCategoryID    string
	deadCategoryID      string
	botRef              *dc.Session
	activeChannelPrefix string
	deadChannelPrefix   string
}

func NewClient(token string, guildID string, activeCategory string, deadCategory string, activeChannelPrefix string, deadChannelPrefix string) *Client {

	c := Client{
		botToken:            token,
		guildID:             guildID,
		activeCategoryID:    activeCategory,
		deadCategoryID:      deadCategory,
		activeChannelPrefix: activeChannelPrefix,
		deadChannelPrefix:   deadChannelPrefix,
	}

	return &c
}

func (c *Client) SendMsg(channelID, content string) (*dc.Message, error) {
	return c.botRef.ChannelMessageSend(channelID, content)
}

func (c *Client) CreateChannel(name string) (*dc.Channel, error) {
	return c.botRef.GuildChannelCreateComplex(c.guildID, dc.GuildChannelCreateData{
		ParentID: c.activeCategoryID,
		Name:     c.activeChannelPrefix + "-" + name,
		Type:     dc.ChannelTypeGuildText,
	})
}

func (c *Client) DeleteChannel(channelID string) (*dc.Channel, error) {
	return c.botRef.ChannelDelete(channelID)
}

func (c *Client) MakeChannelActive(channelID string) (*dc.Channel, error) {
	channel, err := c.botRef.Channel(channelID)
	if err != nil {
		return nil, err
	}

	// Remove dead prefix if present and add active prefix
	name := channel.Name
	if len(c.deadChannelPrefix) > 0 && len(name) > len(c.deadChannelPrefix)+1 {
		name = name[len(c.deadChannelPrefix)+1:]
	}

	return c.botRef.ChannelEdit(channelID, &dc.ChannelEdit{
		ParentID: c.activeCategoryID,
		Name:     c.activeChannelPrefix + "-" + name,
	})
}

func (c *Client) MakeChannelDead(channelID string) (*dc.Channel, error) {
	channel, err := c.botRef.Channel(channelID)
	if err != nil {
		return nil, err
	}

	// Remove active prefix if present and add dead prefix
	name := channel.Name
	if len(c.activeChannelPrefix) > 0 && len(name) > len(c.activeChannelPrefix)+1 {
		name = name[len(c.activeChannelPrefix)+1:]
	}

	return c.botRef.ChannelEdit(channelID, &dc.ChannelEdit{
		ParentID: c.deadCategoryID,
		Name:     c.deadChannelPrefix + "-" + name,
	})
}
