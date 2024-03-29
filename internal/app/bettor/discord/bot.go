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
	Ctx     context.Context
	D       *discordgo.Session
	Client  bettorClient
	Logger  log.Logger
	CleanUp bool

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
func New(ctx context.Context, args ...Arg) (*Bot, error) {
	discordArgs := &discordArgs{
		logger: log.NewNopLogger(),
	}
	for _, arg := range args {
		arg(discordArgs)
	}
	if discordArgs.token == "" || discordArgs.bettorClient == nil || discordArgs.logger == nil {
		return nil, fmt.Errorf("missing required arguments")
	}

	d, err := discordgo.New("Bot " + discordArgs.token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}
	d.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{
		Ctx:      ctx,
		D:        d,
		Client:   discordArgs.bettorClient,
		Logger:   discordArgs.logger,
		CleanUp:  discordArgs.cleanUp,
		Commands: initCommands(ctx, discordArgs.bettorClient, discordArgs.logger),
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

type discordArgs struct {
	token        string
	bettorClient bettorv1alphaconnect.BettorServiceClient
	logger       log.Logger
	cleanUp      bool
}

// Arg is an argument for constructing a Discord bot.
type Arg func(o *discordArgs)

// WithToken provides a Discord token to the Discord bot.
func WithToken(token string) Arg {
	return Arg(func(a *discordArgs) {
		a.token = token
	})
}

// WithBettorClient provides a Bettor client to the Discord bot.
func WithBettorClient(c bettorv1alphaconnect.BettorServiceClient) Arg {
	return Arg(func(a *discordArgs) {
		a.bettorClient = c
	})
}

// WithLogger provides a logger to the Discord bot.
func WithLogger(logger log.Logger) Arg {
	return Arg(func(a *discordArgs) {
		a.logger = logger
	})
}

// WithCleanUp configured Discord bot to clean up registered commands on shutdown.
func WithCleanUp() Arg {
	return Arg(func(a *discordArgs) {
		a.cleanUp = true
	})
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

	if !b.CleanUp {
		b.Logger.Log("msg", "no command cleanup")
		return
	}

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

	defs := make([]*discordgo.ApplicationCommand, len(b.Commands))
	i := 0
	for _, v := range b.Commands {
		defs[i] = v.Def
		i++
	}
	if _, err := s.ApplicationCommandBulkOverwrite(botDiscordUser, guildID, defs); err != nil {
		logger.Log("msg", "failed to create commands", "err", err)
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
