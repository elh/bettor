// Package discord implements a Discord bot that can join and run in guilds.
package discord

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/go-kit/log"
)

// Bot is a Discord Bot for Bettor. Only one instance can be running.
type Bot struct {
	D      *discordgo.Session
	Client bettorv1alphaconnect.BettorServiceClient
	Logger log.Logger

	Commands map[string]*CommandAndHandler

	// keep track of all guilds we have joined so we can clean up commands on termination.
	GuildIDs   []string
	guildIDMtx sync.Mutex
}

// CommandAndHandler is a helper struct for an ApplicationCommand and its handler function.
type CommandAndHandler struct {
	Command *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// New initializes a new Bot.
func New(token string, bettorClient bettorv1alphaconnect.BettorServiceClient, logger log.Logger) (*Bot, error) {
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	d.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{
		D:        d,
		Client:   bettorClient,
		Logger:   logger,
		Commands: nil, // manually set up w/ backwards reference to Bot...
	}
	// source of truth mapping of commands. ensure names are correct by defering to map key here.
	b.Commands = map[string]*CommandAndHandler{
		"start-bet": {
			Command: startBetCommand,
			Handler: b.StartBet,
		},
	}
	for k, v := range b.Commands {
		k := k
		v.Command.Name = k
		handleFn := v.Handler
		v.Handler = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			logger.Log("msg", "received command", "command", k, "user", i.Member.User.ID, "guild", i.GuildID)
			handleFn(s, i)
			// TODO: log success and failure when our command interface is explicit
		}
	}

	// set up handlers
	d.AddHandler(b.guildCreate)
	d.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := b.Commands[i.ApplicationCommandData().Name]; ok {
			h.Handler(s, i)
		}
	})

	return b, nil
}

// Run starts the bot. This blocks until the bot is terminated.
func (b *Bot) Run(ctx context.Context) error {
	// Open websocket and begin listening
	if err := b.D.Open(); err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	defer b.D.Close()
	defer b.cleanup()

	<-ctx.Done()
	return nil
}

// Cleanup upon bot stopping. Deletes all app commands created in all guilds joined.
func (b *Bot) cleanup() {
	b.guildIDMtx.Lock()
	defer b.guildIDMtx.Unlock()

	b.Logger.Log("msg", "cleaning up commands...")
	for _, guildID := range b.GuildIDs {
		cmds, err := b.D.ApplicationCommands(b.D.State.User.ID, guildID)
		if err != nil {
			b.Logger.Log("msg", "failed to get commands for cleanup", "guildID", guildID, "err", err)
			continue
		}
		for _, v := range cmds {
			err := b.D.ApplicationCommandDelete(b.D.State.User.ID, guildID, v.ID)
			if err != nil {
				b.Logger.Log("msg", "failed to delete command for cleanup", "guildID", guildID, "command", v.Name, "err", err)
				continue
			}
		}
	}
}

// when a guild is joined, add command and track it to the list of guilds to clean up on exit.
func (b *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	b.guildIDMtx.Lock()
	defer b.guildIDMtx.Unlock()

	b.Logger.Log("msg", "joining guild", "guildID", event.Guild.ID)
	if event.Guild.Unavailable {
		b.Logger.Log("msg", "guild is unavailable", "guildID", event.Guild.ID)
		return
	}

	for _, v := range b.Commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, event.Guild.ID, v.Command)
		if err != nil {
			b.Logger.Log("msg", "failed to create command", "guildID", event.Guild.ID, "command", v.Command.Name, "err", err)
		}
	}

	b.GuildIDs = append(b.GuildIDs, event.Guild.ID)
	// TODO: sendWelcomeMessage. use when we have a better way to only send on first join to guild.
}

//nolint:deadcode,unused
func (b *Bot) sendWelcomeMessage(s *discordgo.Session, guild *discordgo.Guild) {
	var firstChannelID string
	lowestPosition := math.MaxInt32
	for _, channel := range guild.Channels {
		if channel.Type == discordgo.ChannelTypeGuildText && channel.Position < lowestPosition {
			firstChannelID = channel.ID
			lowestPosition = channel.Position
		}
	}
	if firstChannelID != "" {
		if _, err := s.ChannelMessageSend(firstChannelID, "🎲 Hi! I'm Bettor, your Discord bookmaker. Type `/start-bet` to get started with your first pool bet. All users start with 10K points."); err != nil {
			b.Logger.Log("msg", "failed to send welcome message", "guildID", guild.ID, "err", err)
		}
	}
}
