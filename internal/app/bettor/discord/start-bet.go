package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

const (
	defaultNewUserCentipoints = 1000000 // 10K points
)

var (
	one = 1

	startBetCommand = &discordgo.ApplicationCommand{
		Name:        "start-bet",
		Description: "Start a new bet. At least 2 outcome options are required",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "bet",
				Description: "The bet? Requires at least 2 outcome options",
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
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome3",
				Description: "Outcome 3",
				MinLength:   &one,
				MaxLength:   1024,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome4",
				Description: "Outcome 4",
				MinLength:   &one,
				MaxLength:   1024,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome5",
				Description: "Outcome 5",
				MinLength:   &one,
				MaxLength:   1024,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "outcome6",
				Description: "Outcome 6",
				MinLength:   &one,
				MaxLength:   1024,
			},
		},
	}
)

// StartBet is the handler for the /start-bet command.
func StartBet(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		discordUserID, options, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		// make sure caller user exists. if not, create a new user.
		bettorUser, err := getUserOrCreateIfNotExist(ctx, client, discordUserID)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup (or create nonexistent) user"}, fmt.Errorf("failed to get or create user: %w", err)
		}

		outcomeKeys := []string{"outcome1", "outcome2", "outcome3", "outcome4", "outcome5", "outcome6"}

		var outcomes []*api.Outcome
		for _, k := range outcomeKeys {
			if option, ok := options[k]; ok {
				outcomes = append(outcomes, &api.Outcome{
					Title: option.StringValue(),
				})
			}
		}
		resp, err := client.CreateMarket(ctx, &connect.Request[api.CreateMarketRequest]{Msg: &api.CreateMarketRequest{Market: &api.Market{
			Title:   options["bet"].StringValue(),
			Creator: bettorUser.GetName(),
			Type: &api.Market_Pool{
				Pool: &api.Pool{
					Outcomes: outcomes,
				},
			},
		}}})
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to start bet"}, fmt.Errorf("failed to create market: %w", err)
		}
		market := resp.Msg.GetMarket()

		msgformat, margs := formatMarket(market)
		msgformat = "🎲 🆕 Type `/join-bet` to join the bet until it is locked.\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}

func getUserOrCreateIfNotExist(ctx context.Context, client bettorClient, discordUserID string) (bettorUser *api.User, err error) {
	getUserResp, err := client.GetUserByUsername(ctx, &connect.Request[api.GetUserByUsernameRequest]{Msg: &api.GetUserByUsernameRequest{Username: discordUserID}})
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			if connectErr.Code() == connect.CodeNotFound {
				createUserResp, err := client.CreateUser(ctx, &connect.Request[api.CreateUserRequest]{Msg: &api.CreateUserRequest{User: &api.User{
					Username:    discordUserID,
					Centipoints: defaultNewUserCentipoints,
				}}})
				if err != nil {
					return nil, fmt.Errorf("failed to create user: %w", err)
				}
				return createUserResp.Msg.GetUser(), nil
			}
			return nil, fmt.Errorf("failed to get user, not CodeNotFound: %w", err)
		}
		return nil, fmt.Errorf("failed to get user, not a *connect.Error: %w", err)
	}
	return getUserResp.Msg.GetUser(), nil
}
