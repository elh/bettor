package discord

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var settleBetCommand = &discordgo.ApplicationCommand{
	Name:        "settle-bet",
	Description: "Settle a bet and pay out winners",
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
			Name:         "winner",
			Description:  "Winning outcome",
			Required:     true,
			MinLength:    &one,
			MaxLength:    1024,
			Autocomplete: true,
		},
	},
}

// SettleBet is the handler for the /settle-bet command.
func SettleBet(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		_, options, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			resp, err := client.SettleMarket(ctx, &connect.Request[api.SettleMarketRequest]{Msg: &api.SettleMarketRequest{
				MarketId: options["bet"].StringValue(),
				Type: &api.SettleMarketRequest_WinnerId{
					WinnerId: options["winner"].StringValue(),
				},
			}})
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to settle bet"}, fmt.Errorf("failed to CreateBet: %w", err)
			}
			market := resp.Msg.GetMarket()
			var winnerTitle string
			for _, outcome := range market.GetPool().GetOutcomes() {
				if outcome.GetId() == options["winner"].StringValue() {
					winnerTitle = outcome.GetTitle()
					break
				}
			}

			msgformat, margs := formatMarket(market)
			msgformat = "🎲 ✅ Bet settled with winner **%s**\n" + msgformat
			margs = append([]interface{}{winnerTitle}, margs...)
			return &discordgo.InteractionResponseData{Content: fmt.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{
				Status:   api.Market_STATUS_BETS_LOCKED,
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
						Value: market.GetId(),
					})
				}
			case options["winner"] != nil && options["winner"].Focused:
				if options["bet"] != nil && options["bet"].StringValue() != "" {
					for _, market := range resp.Msg.GetMarkets() {
						if market.GetId() != options["bet"].StringValue() {
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
			return &discordgo.InteractionResponseData{Choices: choices}, nil
		default:
			return &discordgo.InteractionResponseData{Content: "🔺 Something went wrong..."}, fmt.Errorf("unexpected event type %v", event.Type)
		}
	}
}
