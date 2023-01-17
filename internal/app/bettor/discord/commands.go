package discord

import (
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
type Handler func(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error)

// initCommands initializes bot commands from a source of truth mapping. It ensure names are correct by defering to map
// key and instruments the handler.
func initCommands(b *Bot) map[string]*DGCommand {
	commands := map[string]*Command{
		"start-bet": {
			Def:     startBetCommand,
			Handler: b.StartBet,
		},
	}

	out := map[string]*DGCommand{}
	for k, v := range commands {
		k := k
		handlerFn := v.Handler
		v.Def.Name = k // make sure key did not drift
		out[k] = &DGCommand{
			Def: v.Def,
			Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
				logger := log.With(b.Logger, "command", k, "interaction", i.ID, "user", i.Member.User.ID, "guild", i.GuildID)
				now := time.Now()
				respData, err := handlerFn(s, i)
				durMS := time.Now().Sub(now).Milliseconds()
				if err != nil {
					if respData == nil {
						respData = &discordgo.InteractionResponseData{
							Content: "🔺 An error occurred while processing your command.",
						}
					}
					logger.Log("msg", "command handler failure", "dur_ms", durMS, "err", err)
				} else {
					logger.Log("msg", "command handler success", "dur_ms", durMS)
				}

				if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: respData,
				}); err != nil {
					logger.Log("msg", "failed to respond to interaction", "err", err)
				}
			},
		}
	}
	return out
}
