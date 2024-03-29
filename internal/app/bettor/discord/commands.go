package discord

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-kit/log"
)

// Command is a helper struct for an discordgo ApplicationCommand and its handler function. This contains a internal
// Handler type which is nice for generic instrumentation and error handling.
type Command struct {
	Def     *discordgo.ApplicationCommand
	Handler Handler
}

// Handler is an InteractionCreate handler. It returns a InteractionResponseData which will serve as our universal interface
// for generic handling.
type Handler func(*discordgo.Session, *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)

// initCommands initializes bot commands from a source of truth mapping. It ensure names are correct by defering to map
// key and instruments the handler.
func initCommands(ctx context.Context, client bettorClient, logger log.Logger) map[string]*DGCommand {
	commands := map[string]*Command{
		"start-bet": {
			Def:     startBetCommand,
			Handler: StartBet(ctx, client),
		},
		"lock-bet": {
			Def:     lockBetCommand,
			Handler: LockBet(ctx, client),
		},
		"settle-bet": {
			Def:     settleBetCommand,
			Handler: SettleBet(ctx, client),
		},
		"cancel-bet": {
			Def:     cancelBetCommand,
			Handler: CancelBet(ctx, client),
		},
		"get-bet": {
			Def:     getBetCommand,
			Handler: GetBet(ctx, client),
		},
		"join-bet": {
			Def:     joinBetCommand,
			Handler: JoinBet(ctx, client),
		},
		"bettor": {
			Def:     getBettorCommand,
			Handler: GetBettor(ctx, client),
		},
		"bettors": {
			Def:     getBettorsCommand,
			Handler: GetBettors(ctx, client),
		},
	}

	out := map[string]*DGCommand{}
	for k, v := range commands {
		k := k
		handlerFn := v.Handler
		v.Def.Name = k // make sure key did not drift
		out[k] = &DGCommand{
			Def: v.Def,
			Handler: func(s *discordgo.Session, event *discordgo.InteractionCreate) {
				logger := log.With(logger, "command", k, "interaction", event.ID, "user", event.Member.User.ID, "guild", event.GuildID)
				now := time.Now()
				respData, err := handlerFn(s, event)
				durMS := time.Now().Sub(now).Milliseconds()
				if err != nil {
					var cErr *CommandError
					if !errors.As(err, &cErr) {
						cErr = CErr("An error occurred while processing your command.", err)
					}
					respData = &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("🔺 %s", cErr.UserMsg),
					}
					logger.Log("msg", "command handler failure", "dur_ms", durMS, "err", cErr)
				} else {
					logger.Log("msg", "command handler success", "dur_ms", durMS)
				}

				// janky handling up here. could also just have commands return the whole *discordgo.InteractionResponse
				respType := discordgo.InteractionResponseChannelMessageWithSource
				if len(respData.Choices) > 0 {
					respType = discordgo.InteractionApplicationCommandAutocompleteResult
				}
				if err := s.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
					Type: respType,
					Data: respData,
				}); err != nil {
					logger.Log("msg", "failed to respond to interaction", "err", err)
				}
			},
		}
	}
	return out
}

// commandArgs is a helper function to extract the Discord user ID and options from an InteractionCreate event.
func commandArgs(event *discordgo.InteractionCreate) (guildID, userID string, options map[string]*discordgo.ApplicationCommandInteractionDataOption, err error) {
	if event.Member == nil || event.Member.User == nil || event.Member.User.ID == "" {
		return "", "", nil, fmt.Errorf("no user provided in interaction event")
	}
	guildID = event.GuildID
	userID = event.Member.User.ID
	options = map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	for _, opt := range event.ApplicationCommandData().Options {
		options[opt.Name] = opt
	}
	return guildID, userID, options, nil
}
