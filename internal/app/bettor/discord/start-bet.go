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

var _ Handler = (&Bot{}).StartBet

var (
	one = 1

	startBetCommand = &discordgo.ApplicationCommand{
		Name:        "start-bet",
		Description: "Start a new pool bet. At least 2 outcome options are required",
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
func (b *Bot) StartBet(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
	discordUserID, options, err := commandArgs(event)
	if err != nil {
		return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
	}

	// make sure caller user exists. if not, create a new user.
	var bettorUserID string
	getUserResp, err := b.Client.GetUserByUsername(context.Background(), &connect.Request[api.GetUserByUsernameRequest]{Msg: &api.GetUserByUsernameRequest{Username: discordUserID}})
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			if connectErr.Code() == connect.CodeNotFound {
				createUserResp, err := b.Client.CreateUser(context.Background(), &connect.Request[api.CreateUserRequest]{Msg: &api.CreateUserRequest{User: &api.User{
					Username:    discordUserID,
					Centipoints: defaultNewUserCentipoints,
				}}})
				if err != nil {
					return &discordgo.InteractionResponseData{Content: "🔺 Failed to create user"}, fmt.Errorf("failed to create user: %w", err)
				}
				bettorUserID = createUserResp.Msg.GetUser().GetId()
			} else {
				return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup user"}, fmt.Errorf("failed to get user: %w", err)
			}
		} else {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup user"}, fmt.Errorf("failed to get user: %w", err)
		}
	} else {
		bettorUserID = getUserResp.Msg.GetUser().GetId()
	}

	var outcomes []*api.Outcome
	for _, k := range []string{"outcome1", "outcome2", "outcome3", "outcome4", "outcome5", "outcome6"} {
		if option, ok := options[k]; ok {
			outcomes = append(outcomes, &api.Outcome{
				Title: option.StringValue(),
			})
		}
	}
	if _, err = b.Client.CreateMarket(context.Background(), &connect.Request[api.CreateMarketRequest]{Msg: &api.CreateMarketRequest{Market: &api.Market{
		Title:   options["title"].StringValue(),
		Creator: bettorUserID,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: outcomes,
			},
		},
	}}}); err != nil {
		return &discordgo.InteractionResponseData{Content: "🔺 Failed to start bet"}, fmt.Errorf("failed to create market: %w", err)
	}

	margs := make([]interface{}, 0, len(options))
	msgformat := "Created! Type `/join-bet` to join the bet until it is locked.\n"
	for k, v := range map[string]string{ // option key -> message format
		"title":    "Bet",
		"outcome1": "Outcome 1",
		"outcome2": "Outcome 2",
		"outcome3": "Outcome 3",
		"outcome4": "Outcome 4",
		"outcome5": "Outcome 5",
		"outcome6": "Outcome 6",
	} {
		if option, ok := options[k]; ok {
			margs = append(margs, v, option.StringValue())
			msgformat += "> %s: %s\n"
		}
	}

	return &discordgo.InteractionResponseData{Content: fmt.Sprintf(msgformat, margs...)}, nil
}
