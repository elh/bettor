package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	one = 1

	startBetCommand = &discordgo.ApplicationCommand{
		Name:        "start-bet",
		Description: "Start a new pool bet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "title",
				Description: "Bet",
				Required:    true,
				MinLength:   &one,
				MaxLength:   1024,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome1",
				Description: "Outcome 1",
				Required:    true,
				MinLength:   &one,
				MaxLength:   1024,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome2",
				Description: "Outcome 2",
				Required:    true,
				MinLength:   &one,
				MaxLength:   1024,
			},
		},
	}
)

// StartBet is the handler for the /start-bet command.
func (b *Bot) StartBet(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	margs := make([]interface{}, 0, len(options))
	msgformat := "Created! Until locked, bets are open with `/join-bet`\n"

	if option, ok := optionMap["title"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Bet: %s\n"
	}
	if option, ok := optionMap["outcome1"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 1: %s\n"
	}
	if option, ok := optionMap["outcome2"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 2: %s\n"
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(msgformat, margs...),
		},
	}); err != nil {
		b.Logger.Log("msg", "failed to register StartBet handler", "error", err)
	}
}
