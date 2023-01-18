package entity

import "regexp"

const (
	userCollection    = "users"
	marketCollection  = "markets"
	outcomeCollection = "outcomes"
)

var (
	userRegex    = regexp.MustCompile(`users\/(.*)`)
	marketRegex  = regexp.MustCompile(`markets\/(.*)`)
	outcomeRegex = regexp.MustCompile(`markets\/(.*)\/outcomes\/(.*)`)
)

// UserN returns a resource name from user ids.
func UserN(userID string) string {
	return userCollection + "/" + userID
}

// UserIDs returns ids from user resource name.
func UserIDs(name string) (userID string) {
	parts := userRegex.FindStringSubmatch(name)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// MarketN returns a resource name from market ids.
func MarketN(marketID string) string {
	return marketCollection + "/" + marketID
}

// MarketIDs returns ids from market resource name.
func MarketIDs(name string) (marketID string) {
	parts := marketRegex.FindStringSubmatch(name)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// OutcomeN returns a resource name from outcome ids.
func OutcomeN(marketID, outcomeID string) string {
	return marketCollection + "/" + marketID + "/" + outcomeCollection + "/" + outcomeID
}

// OutcomeIDs returns ids from outcome resource name.
func OutcomeIDs(name string) (marketID, outcomeID string) {
	parts := outcomeRegex.FindStringSubmatch(name)
	if len(parts) != 3 {
		return "", ""
	}
	return parts[1], parts[2]
}
