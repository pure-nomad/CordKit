# CordKit

A portable bridge between your C2 infrastructure and Discord, built for quick and lightweight operations.

## Features

### 🟢 Logging
- Automatic transcript channel creation for logging your operations.
- Info & Error logging with timestamps & emojis :3

### 🟢 Custom Commands
- Dynamic slash command registration with a built in management system.
- Default Commands:
  - `/start` - Enable Bot
  - `/stop` - Disable Bot
  - `/purge` - Clean up dead session channels
  - `/nuke` - Complete server cleanup and stop the server process

### 🟢 Session Management
- Automatic channel creation for new sessions
- Archives dead sessions in a seperate category
- Configurable channel naming conventions for your active and dead connections :3

## Configuration

CordKit requires the following configuration parameters:
- Discord Bot Token
- The Guild ID of your Server
- Category IDs (Active Connections, Dead Connections, Transcripts)
- Channel Prefixes (Active, Dead)

## TODO

1. Integrate with Stellarlink
2. Integrate with larger projects:
   - Sliver
   - Ligolo

## Usage

```go
// Basic Setup
clientSettings := cordkit.NewClient(
    "BOT_TOKEN", // Discord Bot Token
    "GUILD_ID", // Discord Server ID
    "ACTIVE_CATEGORY_ID", // ID of your active category
    "DEAD_CATEGORY_ID", // ID of your dead category
    "TRANSCRIPT_CATEGORY_ID", // ID of your transcript category
    "active", // Prefix for active connections
    "dead", // Prefix for dead connections
)

// Start bot with client settings & logging enabled
bot := cordkit.NewBot(clientSettings, true)
bot.CustomCommands = true

// Add a custom command
bot.Commands = append(bot.Commands, cordkit.Command{
    Name:        "ping",
    Description: "Responds with pong",
    Action: func(b *cordkit.Bot, i *dc.InteractionCreate) {
        b.SendMsg(i.ChannelID, "Pong!")
    },
})

bot.Start()
```