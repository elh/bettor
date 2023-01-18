package discord

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var getBettorCommand = &discordgo.ApplicationCommand{
	Name:        "bettor",
	Description: "Get your bettor stats",
}

// GetBettor is the handler for the /get-bet command.
func GetBettor(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		discordUserID, _, err := commandArgs(event)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to handle command"}, fmt.Errorf("failed to handle command: %w", err)
		}

		bettorUser, err := getUserOrCreateIfNotExist(ctx, client, discordUserID)
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to lookup (or create nonexistent) user"}, fmt.Errorf("failed to get or create user: %w", err)
		}

		resp, err := client.ListBets(ctx, &connect.Request[api.ListBetsRequest]{Msg: &api.ListBetsRequest{
			User:           bettorUser.GetName(),
			ExcludeSettled: true,
		}})
		if err != nil {
			return &discordgo.InteractionResponseData{Content: "🔺 Failed to list bets"}, fmt.Errorf("failed to list bets: %w", err)
		}
		var unsettledCentipoints uint64
		for _, b := range resp.Msg.GetBets() {
			unsettledCentipoints += b.GetCentipoints()
		}

		msgformat, margs := formatUser(bettorUser, unsettledCentipoints)
		msgformat = "🎲 👤\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}
