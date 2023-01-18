package entity

import "regexp"

const (
	bookCollection    = "books"
	userCollection    = "users"
	marketCollection  = "markets"
	outcomeCollection = "outcomes"
	betCollection     = "bets"
)

var (
	userRegex    = regexp.MustCompile(`books\/([^/]*)\/users\/([^/]*)`)
	marketRegex  = regexp.MustCompile(`markets\/([^/]*)`)
	outcomeRegex = regexp.MustCompile(`markets\/([^/]*)\/outcomes\/([^/]*)`)
	betRegex     = regexp.MustCompile(`bets\/([^/]*)`)
)

// UserN returns a resource name from user ids.
func UserN(bookID, userID string) string {
	return bookCollection + "/" + bookID + "/" + userCollection + "/" + userID
}

// UserIDs returns ids from user resource name.
func UserIDs(name string) (bookID, userID string) {
	parts := userRegex.FindStringSubmatch(name)
	if len(parts) != 3 {
		return "", ""
	}
	return parts[1], parts[2]
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

// BetN returns a resource name from bet ids.
func BetN(betID string) string {
	return betCollection + "/" + betID
}

// BetIDs returns ids from bet resource name.
func BetIDs(name string) (betID string) {
	parts := betRegex.FindStringSubmatch(name)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
