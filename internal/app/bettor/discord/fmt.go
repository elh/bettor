package discord

import (
	"strings"

	api "github.com/elh/bettor/api/bettor/v1alpha"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var localized = message.NewPrinter(language.English)

// formatUser formats a user for display in Discord.
func formatUser(user *api.User) (fmtStr string, args []interface{}) {
	margs := []interface{}{float32(user.GetCentipoints()) / 100}
	msgformat := "Points: **%v**\n"
	return msgformat, margs
}

// formatMarket formats a market for display in Discord.
func formatMarket(market *api.Market) (fmtStr string, args []interface{}) {
	var totalCentipoints uint64
	for _, outcome := range market.GetPool().GetOutcomes() {
		totalCentipoints += outcome.GetCentipoints()
	}
	margs := []interface{}{market.GetTitle(), strings.TrimPrefix(market.GetStatus().String(), "STATUS_")}
	msgformat := "Bet: **%s**\nStatus: `%s`\n"
	for _, outcome := range market.GetPool().GetOutcomes() {
		if outcome.GetCentipoints() > 0 && totalCentipoints != outcome.GetCentipoints() {
			margs = append(margs, outcome.GetTitle(), (float32(outcome.GetCentipoints()) / 100), float32(totalCentipoints)/float32(outcome.GetCentipoints()))
			msgformat += "- **%s** (Points: **%v**, Odds: **1:%.3f**)"
		} else {
			margs = append(margs, outcome.GetTitle(), float32(outcome.GetCentipoints())/100)
			msgformat += "- **%s** (Points: **%v**, Odds: **-**)"
		}
		if market.GetPool().GetWinnerId() != "" && outcome.GetId() == market.GetPool().GetWinnerId() {
			msgformat += " ✅ "
		}
		msgformat += "\n"
	}
	return msgformat, margs
}

// func humanized(number interface{}) string {
// 	p := message.NewPrinter(language.English)
// 	return p.Sprintf("%v", number)
// }
