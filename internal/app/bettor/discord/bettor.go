package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

var getBettorCommand = &discordgo.ApplicationCommand{
	Name:        "bettor",
	Description: "Get your bettor stats",
}

// GetBettor is the handler for the /bettor command.
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

		msgformat, margs := formatUser(bettorUser, bettorUser.UnsettledCentipoints)
		msgformat = "🎲 👤\n\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}
