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
		guildID, discordUserID, options, err := commandArgs(event)
		if err != nil {
			return nil, CErr("Failed to handle command", err)
		}

		switch event.Type { //nolint:exhaustive
		case discordgo.InteractionApplicationCommand:
			bettorUser, err := getUserOrCreateIfNotExist(ctx, client, guildID, discordUserID)
			if err != nil {
				return nil, CErr("Failed to get or create new user", err)
			}
			bettorUserN := bettorUser.GetName()

			if _, err := client.CreateBet(ctx, &connect.Request[api.CreateBetRequest]{Msg: &api.CreateBetRequest{
				Book: bookName(guildID),
				Bet: &api.Bet{
					User:        bettorUserN,
					Market:      options["bet"].StringValue(),
					Centipoints: uint64(options["points"].FloatValue() * 100),
					Type: &api.Bet_Outcome{
						Outcome: options["outcome"].StringValue(),
					},
				},
			}}); err != nil {
				return nil, CErr("Failed to join bet", err)
			}

			resp, err := client.GetMarket(ctx, &connect.Request[api.GetMarketRequest]{Msg: &api.GetMarketRequest{Name: options["bet"].StringValue()}})
			if err != nil {
				return nil, CErr("Failed to lookup bet", err)
			}
			market := resp.Msg.GetMarket()
			var outcomeTitle string
			for _, outcome := range market.GetPool().GetOutcomes() {
				if outcome.GetName() == options["outcome"].StringValue() {
					outcomeTitle = outcome.GetTitle()
					break
				}
			}

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
			msgformat = "🎲 🪙 <@!%s> bet **%v** points on **%s**.\n\n" + msgformat
			margs = append([]interface{}{discordUserID, options["points"].FloatValue(), outcomeTitle}, margs...)
			return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
		case discordgo.InteractionApplicationCommandAutocomplete:
			resp, err := client.ListMarkets(ctx, &connect.Request[api.ListMarketsRequest]{Msg: &api.ListMarketsRequest{
				Book:     bookName(guildID),
				Status:   api.Market_STATUS_OPEN,
				PageSize: 25,
			}})
			if err != nil {
				return nil, CErr("Failed to lookup bets", err)
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
								Value: outcome.GetName(),
							})
						}
					}
				}
			}
			return &discordgo.InteractionResponseData{Choices: withDefaultChoices(choices)}, nil
		default:
			return nil, CErr("Something went wrong", fmt.Errorf("unexpected event type %v", event.Type))
		}
	}
}
