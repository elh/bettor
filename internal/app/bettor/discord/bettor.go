package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
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

		msgformat, margs := formatUser(bettorUser)
		msgformat = "🎲 👤\n" + msgformat
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, margs...)}, nil
	}
}
