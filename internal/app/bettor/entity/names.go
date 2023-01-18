package entity

import "regexp"

const userCollection = "users"

var userRegex = regexp.MustCompile(`users\/(.*)`)

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
