package discord

import (
	"context"
	"fmt"

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
		guildID, _, options, err := commandArgs(event)
		if err != nil {
			return nil, CErr("Failed to handle command", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			resp, err := client.GetMarket(ctx, &connect.Request[api.GetMarketRequest]{Msg: &api.GetMarketRequest{Name: options["bet"].StringValue()}})
			if err != nil {
				return nil, CErr("Failed to lookup bet", err)
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
			msgformat = "🎲\n" + msgformat
			return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			var choices []*discordgo.ApplicationCommandOptionChoice
			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{Book: bookName(guildID), PageSize: 25}})
			if err != nil {
				return nil, CErr("Failed to lookup bets", err)
			}
			for _, market := range resp.Msg.GetMarkets() {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  market.GetTitle(),
					Value: market.GetName(),
				})
			}
			return &discordgo.InteractionResponseData{Choices: withDefaultChoices(choices)}, nil
		default:
			return nil, CErr("Something went wrong", fmt.Errorf("unexpected event type %v", event.Type))
		}
	}
}

func withDefaultChoices(choices []*discordgo.ApplicationCommandOptionChoice) []*discordgo.ApplicationCommandOptionChoice {
	if len(choices) == 0 {
		return []*discordgo.ApplicationCommandOptionChoice{
			{Name: "❌ None", Value: ""},
		}
	}
	return choices
}
