package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/bwmarrin/discordgo"
	api "github.com/elh/bettor/api/bettor/v1alpha"
)

var getBettorsCommand = &discordgo.ApplicationCommand{
	Name:        "bettors",
	Description: "Get bettor leaderboard",
}

// GetBettors is the handler for the /get-bet command.
func GetBettors(ctx context.Context, client bettorClient) Handler {
	return func(s *discordgo.Session, event *discordgo.InteractionCreate) (*discordgo.InteractionResponseData, error) {
		guildID, _, _, err := commandArgs(event)
		if err != nil {
			return nil, CErr("Failed to handle command", err)
		}

		listUsersResp, err := client.ListUsers(ctx, &connect.Request[api.ListUsersRequest]{Msg: &api.ListUsersRequest{
			Book:     guildBookName(guildID),
			PageSize: 10,
			OrderBy:  "total_centipoints",
		}})
		if err != nil {
			return nil, CErr("Failed to lookup bettors", err)
		}

		var formattedUsers []string
		var args []interface{}
		for i, user := range listUsersResp.Msg.GetUsers() {
			msgformat, margs := formatUser(user, user.UnsettledCentipoints)
			// hack
			msgformat = strings.TrimSuffix(msgformat, "\n")

			msgformat = fmt.Sprintf("%d", i+1) + ") " + msgformat
			switch i {
			case 0:
				msgformat += " 🥇"
			case 1:
				msgformat += " 🥈"
			case 2:
				msgformat += " 🥉"
			}

			formattedUsers = append(formattedUsers, msgformat)
			args = append(args, margs...)
		}

		msgformat := "🎲 👤\n\n" + strings.Join(formattedUsers, "\n")
		return &discordgo.InteractionResponseData{Content: localized.Sprintf(msgformat, args...)}, nil
	}
}
