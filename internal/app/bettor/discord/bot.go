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

type bettorClient bettorv1alphaconnect.BettorServiceClient

// Bot is a Discord Bot for Bettor. Only one instance can be running.
type Bot struct {
	Ctx    context.Context
	D      *discordgo.Session
	Client bettorClient
	Logger log.Logger

	Commands map[string]*DGCommand

	// keep track of all guilds we have joined so we can clean up commands on termination.
	GuildIDs   []string
	guildIDMtx sync.Mutex
}

// DGCommand is a helper struct for an discordgo ApplicationCommand and its handler function. This contains types ready
// to be used with a discordgo.Session.
type DGCommand struct {
	Def     *discordgo.ApplicationCommand
	Handler func(*discordgo.Session, *discordgo.InteractionCreate)
}

// New initializes a new Bot.
func New(ctx context.Context, token string, bettorClient bettorv1alphaconnect.BettorServiceClient, logger log.Logger) (*Bot, error) {
	d, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	d.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{
		Ctx:      ctx,
		D:        d,
		Client:   bettorClient,
		Logger:   logger,
		Commands: initCommands(ctx, bettorClient, logger),
	}

	// set up handlers
	d.AddHandler(b.guildCreate)
	d.AddHandler(func(s *discordgo.Session, event *discordgo.InteractionCreate) {
		if h, ok := b.Commands[event.ApplicationCommandData().Name]; ok {
			h.Handler(s, event)
		}
	})

	return b, nil
}

// Run starts the bot. This blocks until the bot is terminated.
func (b *Bot) Run() error {
	// Open websocket and begin listening
	if err := b.D.Open(); err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	defer b.D.Close()
	defer b.cleanup()

	<-b.Ctx.Done()
	return nil
}

// Cleanup upon bot stopping. Deletes all app commands created in all guilds joined.
func (b *Bot) cleanup() {
	b.guildIDMtx.Lock()
	defer b.guildIDMtx.Unlock()

	if b.D == nil || b.D.State == nil || b.D.State.User == nil {
		b.Logger.Log("msg", "failed to get bot user for cleanup")
		return
	}
	b.Logger.Log("msg", "cleaning up commands...")
	botDiscordUser := b.D.State.User.ID
	for _, guildID := range b.GuildIDs {
		cmds, err := b.D.ApplicationCommands(botDiscordUser, guildID)
		if err != nil {
			b.Logger.Log("msg", "failed to get commands for cleanup", "guildID", guildID, "err", err)
			continue
		}
		for _, v := range cmds {
			if err := b.D.ApplicationCommandDelete(botDiscordUser, guildID, v.ID); err != nil {
				b.Logger.Log("msg", "failed to delete command for cleanup", "guildID", guildID, "command", v.Name, "err", err)
				continue
			}
		}
	}
}

// when a guild is joined, register commands and track it to the list of guilds to clean up on exit.
func (b *Bot) guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	b.guildIDMtx.Lock()
	defer b.guildIDMtx.Unlock()

	guildID := event.Guild.ID
	botDiscordUser := s.State.User.ID
	logger := log.With(b.Logger, "guildID", guildID)

	if event.Guild.Unavailable {
		logger.Log("msg", "added to unavailable guild, skipping")
		return
	}

	for _, v := range b.Commands {
		_, err := s.ApplicationCommandCreate(botDiscordUser, guildID, v.Def)
		if err != nil {
			logger.Log("msg", "failed to create command", "command", v.Def.Name, "err", err)
		}
	}
	logger.Log("msg", "joined guild")

	b.GuildIDs = append(b.GuildIDs, guildID)
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
		if _, err := s.ChannelMessageSend(firstChannelID, "🎲 Hi! I'm Bettor, your Discord bookmaker. Type `/start-bet` to get started with your first bet. All users start with 10K points."); err != nil {
			b.Logger.Log("msg", "failed to send welcome message", "guildID", guild.ID, "err", err)
		}
	}
}
