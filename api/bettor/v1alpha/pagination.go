package bettorv1alpha

import "google.golang.org/protobuf/proto"

// StripListUsersPagination assists verifying next page tokens
func StripListUsersPagination(in *ListUsersRequest) *ListUsersRequest {
	out := proto.Clone(in).(*ListUsersRequest)
	out.PageSize = 0
	out.PageToken = ""
	return out
}
