package entity

import "regexp"

const (
	userCollection   = "users"
	marketCollection = "markets"
)

var (
	userRegex   = regexp.MustCompile(`users\/(.*)`)
	marketRegex = regexp.MustCompile(`markets\/(.*)`)
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
