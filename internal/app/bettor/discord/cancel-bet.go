package discord

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var cancelBetCommand = &discordgo.ApplicationCommand{
	Name:        "cancel-bet",
	Description: "Cancel a locked bet and refund bettors. Only the bet creator can cancel the bet",
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

// CancelBet is the handler for the /cancel-bet command.
func CancelBet(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		_, _, options, err := commandArgs(event)
		if err != nil {
			return nil, CErr("Failed to handle command", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			resp, err := client.CancelMarket(ctx, &connect.Request[api.CancelMarketRequest]{Msg: &api.CancelMarketRequest{
				Name: options["bet"].StringValue(),
			}})
			if err != nil {
				return nil, CErr("Failed to cancel bet", err)
			}
			market := resp.Msg.GetMarket()

			userResp, err := client.GetUser(ctx, &connect.Request[api.GetUserRequest]{Msg: &api.GetUserRequest{Name: market.GetCreator()}})
			if err != nil {
				return nil, CErr("Failed to lookup bet creator", err)
			}
			marketCreator := userResp.Msg.GetUser()

			bets, bettors, err := getMarketBets(ctx, client, market.GetName())
			if err != nil {
				return nil, CErr("Failed to lookup bettors", err)
			}

			msgformat, margs := formatMarket(market, marketCreator, bets, bettors)
			msgformat = "🎲 ❌ Bet canceled and bettors refunded.\n\n" + msgformat
			return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			guildID, discordUserID, _, err := commandArgs(event)
			if err != nil {
				return nil, CErr("Failed to handle command", err)
			}
			bettorUser, err := getUserOrCreateIfNotExist(ctx, client, guildID, discordUserID)
			if err != nil {
				return nil, CErr("Failed to lookup (or create nonexistent) user", err)
			}

			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{
				Book:     guildBookName(guildID),
				Status:   api.Market_STATUS_BETS_LOCKED,
				PageSize: 25,
			}})
			if err != nil {
				return nil, CErr("Failed to lookup bets", err)
			}

			var choices []*discordgo.ApplicationCommandOptionChoice
			if options["bet"] != nil && options["bet"].Focused {
				for _, market := range resp.Msg.GetMarkets() {
					if market.GetCreator() != bettorUser.GetName() {
						continue
					}
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  market.GetTitle(),
						Value: market.GetName(),
					})
				}
			}
			return &discordgo.InteractionResponseData{Choices: withDefaultChoices(choices)}, nil
		default:
			return nil, CErr("Something went wrong", fmt.Errorf("unexpected event type %v", event.Type))
		}
	}
}
