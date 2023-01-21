package discord

import (
	"context"

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
		guildID, discordUserID, _, err := commandArgs(event)
		if err != nil {
			return nil, CErr("Failed to handle command", err)
		}

		bettorUser, err := getUserOrCreateIfNotExist(ctx, client, guildID, discordUserID)
		if err != nil {
			return nil, CErr("Failed to get or create new user", err)
		}

		resp, err := client.ListBets(ctx, &connect.Request[api.ListBetsRequest]{Msg: &api.ListBetsRequest{
			Book:           bookName(guildID),
			User:           bettorUser.GetName(),
			ExcludeSettled: true,
		}})
		if err != nil {
			return nil, CErr("Failed to list bets", err)
		}
		var unsettledCentipoints uint64
		for _, b := range resp.Msg.GetBets() {
			unsettledCentipoints += b.GetCentipoints()
		}

		msgformat, margs := formatUser(bettorUser, unsettledCentipoints)
		msgformat = "🎲 👤\n\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}
