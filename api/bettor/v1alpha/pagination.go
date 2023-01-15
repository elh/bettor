package bettorv1alpha

import "google.golang.org/protobuf/proto"

// StripListUsersPagination assists verifying next page tokens.
func StripListUsersPagination(in *ListUsersRequest) *ListUsersRequest {
	out := proto.Clone(in).(*ListUsersRequest)
	out.PageSize = 0
	out.PageToken = ""
	return out
}

// StripListMarketsPagination assists verifying next page tokens.
func StripListMarketsPagination(in *ListMarketsRequest) *ListMarketsRequest {
	out := proto.Clone(in).(*ListMarketsRequest)
	out.PageSize = 0
	out.PageToken = ""
	return out
}

// StripListBetsPagination assists verifying next page tokens.
func StripListBetsPagination(in *ListBetsRequest) *ListBetsRequest {
	out := proto.Clone(in).(*ListBetsRequest)
	out.PageSize = 0
	out.PageToken = ""
	return out
}
