package discord

import (
	"strings"

	api "github.com/elh/bettor/api/bettor/v1alpha"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var localized = message.NewPrinter(language.English)

// formatUser formats a user for display in Discord.
func formatUser(user *api.User, unsettledCentipoints uint64) (fmtStr string, args []interface{}) {
	margs := []interface{}{user.GetUsername(), float32(user.GetCentipoints()) / 100, float32(unsettledCentipoints) / 100}
	msgformat := "<@!%s> Points: **%v** (Unsettled points: **%v**)\n"
	return msgformat, margs
}

// formatMarket formats a market for display in Discord.
func formatMarket(market *api.Market, creator *api.User, bets []*api.Bet, bettors []*api.User) (fmtStr string, args []interface{}) {
	var totalCentipoints uint64
	for _, outcome := range market.GetPool().GetOutcomes() {
		totalCentipoints += outcome.GetCentipoints()
	}
	margs := []interface{}{market.GetTitle(), creator.GetUsername(), strings.TrimPrefix(market.GetStatus().String(), "STATUS_")}
	msgformat := "Bet: **%s**\nCreator: <@!%s>\nStatus: `%s`\n"
	for _, outcome := range market.GetPool().GetOutcomes() {
		if outcome.GetCentipoints() > 0 && totalCentipoints != outcome.GetCentipoints() {
			margs = append(margs, outcome.GetTitle(), (float32(outcome.GetCentipoints()) / 100), float32(totalCentipoints)/float32(outcome.GetCentipoints()))
			msgformat += "- **%s** (Points: **%v**, Odds: **1:%.3f**)"
		} else {
			margs = append(margs, outcome.GetTitle(), float32(outcome.GetCentipoints())/100)
			msgformat += "- **%s** (Points: **%v**, Odds: **-**)"
		}

		outcomeBettors := map[string]bool{} // user resource names
		for _, bet := range bets {
			if bet.GetOutcome() != outcome.GetName() {
				continue
			}
			for _, bettor := range bettors {
				if bettor.GetName() != bet.GetUser() {
					continue
				}
				if outcomeBettors[bettor.GetName()] {
					continue
				}
				outcomeBettors[bettor.GetName()] = true
				margs = append(margs, bettor.GetUsername())
				msgformat += " <@!%s>"
			}
		}

		if market.GetPool().GetWinner() != "" && outcome.GetName() == market.GetPool().GetWinner() {
			msgformat += " ✅ "
		}
		msgformat += "\n"
	}
	return msgformat, margs
}
