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
func (b *Bot) StartBet(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	// make sure caller user exists. if not, create a new user.
	var bettorUserID string
	getUserResp, err := b.Client.GetUserByUsername(context.Background(), &connect.Request[api.GetUserByUsernameRequest]{Msg: &api.GetUserByUsernameRequest{Username: i.Member.User.ID}})
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			if connectErr.Code() == connect.CodeNotFound {
				createUserResp, err := b.Client.CreateUser(context.Background(), &connect.Request[api.CreateUserRequest]{Msg: &api.CreateUserRequest{User: &api.User{
					Username:    i.Member.User.ID,
					Centipoints: defaultNewUserCentipoints,
				}}})
				if err != nil {
					b.Logger.Log("msg", "failed to create user", "error", err)
					return
				}
				bettorUserID = createUserResp.Msg.GetUser().GetId()
			} else {
				b.Logger.Log("msg", "failed to get user", "error", err)
				return
			}
		} else {
			b.Logger.Log("msg", "failed to get user", "error", err)
			return
		}
	} else {
		bettorUserID = getUserResp.Msg.GetUser().GetId()
	}

	var outcomes []*api.Outcome
	for _, k := range []string{"outcome1", "outcome2", "outcome3", "outcome4", "outcome5", "outcome6"} {
		if option, ok := optionMap[k]; ok {
			outcomes = append(outcomes, &api.Outcome{
				Title: option.StringValue(),
			})
		}
	}
	_, err = b.Client.CreateMarket(context.Background(), &connect.Request[api.CreateMarketRequest]{Msg: &api.CreateMarketRequest{Market: &api.Market{
		Title:   optionMap["title"].StringValue(),
		Creator: bettorUserID,
		Type: &api.Market_Pool{
			Pool: &api.Pool{
				Outcomes: outcomes,
			},
		},
	}}})
	if err != nil {
		b.Logger.Log("msg", "failed to create market", "error", err)
		return
	}

	margs := make([]interface{}, 0, len(options))
	msgformat := "Created! Type `/join-bet` to join the bet until it is locked.\n"

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
	if option, ok := optionMap["outcome3"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 3: %s\n"
	}
	if option, ok := optionMap["outcome4"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 4: %s\n"
	}
	if option, ok := optionMap["outcome5"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 5: %s\n"
	}
	if option, ok := optionMap["outcome6"]; ok {
		margs = append(margs, option.StringValue())
		msgformat += "> Outcome 6: %s\n"
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(msgformat, margs...),
		},
	}); err != nil {
		b.Logger.Log("msg", "failed to InteractionRespond to StartBet", "error", err)
	}
}
