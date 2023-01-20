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
	bookRegex    = regexp.MustCompile(`books\/([^/]*)`)
	userRegex    = regexp.MustCompile(`books\/([^/]*)\/users\/([^/]*)`)
	marketRegex  = regexp.MustCompile(`books\/([^/]*)\/markets\/([^/]*)`)
	outcomeRegex = regexp.MustCompile(`books\/([^/]*)\/markets\/([^/]*)\/outcomes\/([^/]*)`)
	betRegex     = regexp.MustCompile(`books\/([^/]*)\/bets\/([^/]*)`)
)

// BookN returns a resource name from book ids.
func BookN(bookID string) string {
	return bookCollection + "/" + bookID
}

// BooksIDs returns ids from book resource name.
func BooksIDs(name string) (bookID string) {
	parts := bookRegex.FindStringSubmatch(name)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

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
func MarketN(bookID, marketID string) string {
	return bookCollection + "/" + bookID + "/" + marketCollection + "/" + marketID
}

// MarketIDs returns ids from market resource name.
func MarketIDs(name string) (bookID, marketID string) {
	parts := marketRegex.FindStringSubmatch(name)
	if len(parts) != 3 {
		return "", ""
	}
	return parts[1], parts[2]
}

// OutcomeN returns a resource name from outcome ids.
func OutcomeN(bookID, marketID, outcomeID string) string {
	return bookCollection + "/" + bookID + "/" + marketCollection + "/" + marketID + "/" + outcomeCollection + "/" + outcomeID
}

// OutcomeIDs returns ids from outcome resource name.
func OutcomeIDs(name string) (bookID, marketID, outcomeID string) {
	parts := outcomeRegex.FindStringSubmatch(name)
	if len(parts) != 4 {
		return "", "", ""
	}
	return parts[1], parts[2], parts[3]
}

// BetN returns a resource name from bet ids.
func BetN(bookID, betID string) string {
	return bookCollection + "/" + bookID + "/" + betCollection + "/" + betID
}

// BetIDs returns ids from bet resource name.
func BetIDs(name string) (bookID, betID string) {
	parts := betRegex.FindStringSubmatch(name)
	if len(parts) != 3 {
		return "", ""
	}
	return parts[1], parts[2]
}
