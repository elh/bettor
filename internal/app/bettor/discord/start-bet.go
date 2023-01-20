package discord

import (
	"context"
	"errors"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/internal/app/bettor/entity"
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
		guildID, discordUserID, options, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		// make sure caller user exists. if not, create a new user.
		bettorUser, err := getUserOrCreateIfNotExist(ctx, client, guildID, discordUserID)
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
		resp, err := client.CreateMarket(ctx, &connect.Request[api.CreateMarketRequest]{Msg: &api.CreateMarketRequest{
			Book: bookName(guildID),
			Market: &api.Market{
				Title:   options["bet"].StringValue(),
				Creator: bettorUser.GetName(),
				Type: &api.Market_Pool{
					Pool: &api.Pool{
						Outcomes: outcomes,
					},
				},
			},
		}})
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to start bet"}, fmt.Errorf("failed to create market: %w", err)
		}
		market := resp.Msg.GetMarket()

		bets, bettors, err := getMarketBets(ctx, client, market.GetName())
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup bettors"}, fmt.Errorf("failed to getMarketBets: %w", err)
		}

		msgformat, margs := formatMarket(market, bettorUser, bets, bettors)
		msgformat = "🎲 🆕 New bet started. `/join-bet` to join the bet until it is locked with `/lock-bet`.\n\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}

func getUserOrCreateIfNotExist(ctx context.Context, client bettorClient, guildID, discordUserID string) (bettorUser *api.User, err error) {
	getUserResp, err := client.GetUserByUsername(ctx, &connect.Request[api.GetUserByUsernameRequest]{Msg: &api.GetUserByUsernameRequest{
		Book:     bookName(guildID),
		Username: discordUserID,
	}})
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			if connectErr.Code() == connect.CodeNotFound {
				createUserResp, err := client.CreateUser(ctx, &connect.Request[api.CreateUserRequest]{Msg: &api.CreateUserRequest{
					Book: bookName(guildID),
					User: &api.User{
						Username:    discordUserID,
						Centipoints: defaultNewUserCentipoints,
					},
				}})
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

func bookName(guildID string) string {
	return fmt.Sprintf("books/discord:%s", guildID)
}

// returns a potentially nonexhaustive list of bettors in a market.
func getMarketBets(ctx context.Context, client bettorClient, marketName string) ([]*api.Bet, []*api.User, error) {
	bookID, _ := entity.MarketIDs(marketName)
	betsResp, err := client.ListBets(ctx, &connect.Request[api.ListBetsRequest]{Msg: &api.ListBetsRequest{
		PageSize: 50,
		Market:   marketName,
	}})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list bets: %w", err)
	}
	bets := betsResp.Msg.GetBets()

	var userIDs []string //nolint:prealloc
	for _, bet := range bets {
		userIDs = append(userIDs, bet.GetUser())
	}
	userResp, err := client.ListUsers(ctx, &connect.Request[api.ListUsersRequest]{Msg: &api.ListUsersRequest{
		Book:     bookName(bookID),
		PageSize: 50,
		Users:    userIDs,
	}})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list users: %w", err)
	}
	return bets, userResp.Msg.GetUsers(), nil
}
