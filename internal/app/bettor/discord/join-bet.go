package discord

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var (
	oneFloat       = 1.0
	joinBetCommand = &discordgo.ApplicationCommand{
		Name:        "join-bet",
		Description: "Join in on a bet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "bet",
				Description:  "Bet",
				Required:     true,
				MinLength:    &one,
				MaxLength:    1024,
				Autocomplete: true,
			},
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "outcome",
				Description:  "Outcome",
				Required:     true,
				MinLength:    &one,
				MaxLength:    1024,
				Autocomplete: true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "points",
				Description: "Points",
				Required:    true,
				MinValue:    &oneFloat,
			},
		},
	}
)

// JoinBet is the handler for the /join-bet command.
func JoinBet(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		discordUserID, options, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			bettorUser, err := getUserOrCreateIfNotExist(ctx, client, discordUserID)
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup (or create nonexistent) user"}, fmt.Errorf("failed to get or create user: %w", err)
			}
			bettorUserN := bettorUser.GetName()

			if _, err := client.CreateBet(ctx, &connect.Request[api.CreateBetRequest]{Msg: &api.CreateBetRequest{
				Bet: &api.Bet{
					User:        bettorUserN,
					Market:      options["bet"].StringValue(),
					Centipoints: uint64(options["points"].FloatValue() * 100),
					Type: &api.Bet_OutcomeId{
						OutcomeId: options["outcome"].StringValue(),
					},
				},
			}}); err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to join bet"}, fmt.Errorf("failed to CreateBet: %w", err)
			}

			resp, err := client.GetMarket(ctx, &connect.Request[api.GetMarketRequest]{Msg: &api.GetMarketRequest{Name: options["bet"].StringValue()}})
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup bet"}, fmt.Errorf("failed to GetMarket: %w", err)
			}
			market := resp.Msg.GetMarket()
			var outcomeTitle string
			for _, outcome := range market.GetPool().GetOutcomes() {
				if outcome.GetId() == options["outcome"].StringValue() {
					outcomeTitle = outcome.GetTitle()
					break
				}
			}

			msgformat, margs := formatMarket(market)
			msgformat = "🎲 🪙 Bet **%v** points on **%s**\n" + msgformat
			margs = append([]interface{}{options["points"].FloatValue(), outcomeTitle}, margs...)
			return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{
				Status:   api.Market_STATUS_OPEN,
				PageSize: 25,
			}})
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup bets"}, fmt.Errorf("failed to ListMarkets: %w", err)
			}

			var choices []*discordgo.ApplicationCommandOptionChoice
			switch {
			case options["bet"] != nil && options["bet"].Focused:
				for _, market := range resp.Msg.GetMarkets() {
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  market.GetTitle(),
						Value: market.GetName(),
					})
				}
			case options["outcome"] != nil && options["outcome"].Focused:
				if options["bet"] != nil && options["bet"].StringValue() != "" {
					for _, market := range resp.Msg.GetMarkets() {
						if market.GetName() != options["bet"].StringValue() {
							continue
						}
						for _, outcome := range market.GetPool().GetOutcomes() {
							choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
								Name:  outcome.GetTitle(),
								Value: outcome.GetId(),
							})
						}
					}
				}
			}
			return &discordgo.InteractionResponseData{Choices: withDefaultChoices(choices)}, nil
		default:
			return &discordgo.InteractionResponseData{Content: "🔺 Something went wrong..."}, fmt.Errorf("unexpected event type %v", event.Type)
		}
	}
}
