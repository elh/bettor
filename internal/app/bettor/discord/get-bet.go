package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var getBetCommand = &discordgo.ApplicationCommand{
	Name:        "get-bet",
	Description: "Get information about a bet",
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
	},
}

// GetBet is the handler for the /get-bet command.
func GetBet(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		_, options, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			resp, err := client.GetMarket(ctx, &connect.Request[api.GetMarketRequest]{Msg: &api.GetMarketRequest{MarketId: options["bet"].StringValue()}})
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup bet"}, fmt.Errorf("failed to GetMarket: %w", err)
			}
			market := resp.Msg.GetMarket()

			msgformat, margs := formatMarket(market)
			msgformat = "🎲\n" + msgformat
			return &discordgo.InteractionResponseData{Content: fmt.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			var choices []*discordgo.ApplicationCommandOptionChoice
			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{PageSize: 25}})
			if err != nil {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup bets"}, fmt.Errorf("failed to ListMarkets: %w", err)
			}
			for _, market := range resp.Msg.GetMarkets() {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  market.GetTitle(),
					Value: market.GetId(),
				})
			}
			return &discordgo.InteractionResponseData{Choices: choices}, nil
		default:
			return &discordgo.InteractionResponseData{Content: "🔺 Something went wrong..."}, fmt.Errorf("unexpected event type %v", event.Type)
		}
	}
}

// formatMarket formats a market for display in Discord.
func formatMarket(market *api.Market) (fmtStr string, args []interface{}) {
	var totalCentipoints uint64
	for _, outcome := range market.GetPool().GetOutcomes() {
		totalCentipoints += outcome.GetCentipoints()
	}
	margs := []interface{}{market.GetTitle(), strings.TrimPrefix(market.GetStatus().String(), "STATUS_")}
	msgformat := "Bet: **%s**\nStatus: **%s**\n"
	for _, outcome := range market.GetPool().GetOutcomes() {
		if outcome.GetCentipoints() > 0 && totalCentipoints != outcome.GetCentipoints() {
			margs = append(margs, outcome.GetTitle(), float32(outcome.GetCentipoints())/100, float32(totalCentipoints)/float32(outcome.GetCentipoints()))
			msgformat += "- **%s** (Points: **%v**, Odds: **1:%.3f**)"
		} else {
			margs = append(margs, outcome.GetTitle(), float32(outcome.GetCentipoints())/100)
			msgformat += "- **%s** (Points: **%v**, Odds: **-**)"
		}
		if market.GetPool().GetWinnerId() != "" && outcome.GetId() == market.GetPool().GetWinnerId() {
			msgformat += " ✅ "
		}
		msgformat += "\n"
	}
	return msgformat, margs
}
