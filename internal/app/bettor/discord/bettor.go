package discord

import (
	"context"
	"fmt"

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

		msgformat, margs := formatUser(bettorUser)
		msgformat = "🎲 👤\n" + msgformat
		return &discordgo.InteractionResponseData{Content: fmt.Sprintf(msgformat, margs...)}, nil
	}
}

// formatUser formats a user for display in Discord.
func formatUser(user *api.User) (fmtStr string, args []interface{}) {
	margs := []interface{}{float32(user.GetCentipoints()) / 100}
	msgformat := "Points: **%v**\n"
	return msgformat, margs
}
