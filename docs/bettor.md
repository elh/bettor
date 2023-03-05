# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [bettor/v1alpha/bettor.proto](#bettor_v1alpha_bettor-proto)
    - [Bet](#bettor-v1alpha-Bet)
    - [CancelMarketRequest](#bettor-v1alpha-CancelMarketRequest)
    - [CancelMarketResponse](#bettor-v1alpha-CancelMarketResponse)
    - [CreateBetRequest](#bettor-v1alpha-CreateBetRequest)
    - [CreateBetResponse](#bettor-v1alpha-CreateBetResponse)
    - [CreateMarketRequest](#bettor-v1alpha-CreateMarketRequest)
    - [CreateMarketResponse](#bettor-v1alpha-CreateMarketResponse)
    - [CreateUserRequest](#bettor-v1alpha-CreateUserRequest)
    - [CreateUserResponse](#bettor-v1alpha-CreateUserResponse)
    - [GetBetRequest](#bettor-v1alpha-GetBetRequest)
    - [GetBetResponse](#bettor-v1alpha-GetBetResponse)
    - [GetMarketRequest](#bettor-v1alpha-GetMarketRequest)
    - [GetMarketResponse](#bettor-v1alpha-GetMarketResponse)
    - [GetUserByUsernameRequest](#bettor-v1alpha-GetUserByUsernameRequest)
    - [GetUserByUsernameResponse](#bettor-v1alpha-GetUserByUsernameResponse)
    - [GetUserRequest](#bettor-v1alpha-GetUserRequest)
    - [GetUserResponse](#bettor-v1alpha-GetUserResponse)
    - [ListBetsRequest](#bettor-v1alpha-ListBetsRequest)
    - [ListBetsResponse](#bettor-v1alpha-ListBetsResponse)
    - [ListMarketsRequest](#bettor-v1alpha-ListMarketsRequest)
    - [ListMarketsResponse](#bettor-v1alpha-ListMarketsResponse)
    - [ListUsersRequest](#bettor-v1alpha-ListUsersRequest)
    - [ListUsersResponse](#bettor-v1alpha-ListUsersResponse)
    - [LockMarketRequest](#bettor-v1alpha-LockMarketRequest)
    - [LockMarketResponse](#bettor-v1alpha-LockMarketResponse)
    - [Market](#bettor-v1alpha-Market)
    - [Outcome](#bettor-v1alpha-Outcome)
    - [Pool](#bettor-v1alpha-Pool)
    - [SettleMarketRequest](#bettor-v1alpha-SettleMarketRequest)
    - [SettleMarketResponse](#bettor-v1alpha-SettleMarketResponse)
    - [User](#bettor-v1alpha-User)
  
    - [Market.Status](#bettor-v1alpha-Market-Status)
  
    - [BettorService](#bettor-v1alpha-BettorService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="bettor_v1alpha_bettor-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## bettor/v1alpha/bettor.proto



<a name="bettor-v1alpha-Bet"></a>

### Bet
A user&#39;s bet on a betting market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| settled_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| user | [string](#string) |  |  |
| market | [string](#string) |  |  |
| centipoints | [uint64](#uint64) |  |  |
| settled_centipoints | [uint64](#uint64) |  |  |
| outcome | [string](#string) |  |  |






<a name="bettor-v1alpha-CancelMarketRequest"></a>

### CancelMarketRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="bettor-v1alpha-CancelMarketResponse"></a>

### CancelMarketResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-CreateBetRequest"></a>

### CreateBetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [string](#string) |  |  |
| bet | [Bet](#bettor-v1alpha-Bet) |  |  |






<a name="bettor-v1alpha-CreateBetResponse"></a>

### CreateBetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bet | [Bet](#bettor-v1alpha-Bet) |  |  |






<a name="bettor-v1alpha-CreateMarketRequest"></a>

### CreateMarketRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [string](#string) |  |  |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-CreateMarketResponse"></a>

### CreateMarketResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-CreateUserRequest"></a>

### CreateUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [string](#string) |  |  |
| user | [User](#bettor-v1alpha-User) |  |  |






<a name="bettor-v1alpha-CreateUserResponse"></a>

### CreateUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#bettor-v1alpha-User) |  |  |






<a name="bettor-v1alpha-GetBetRequest"></a>

### GetBetRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bet | [string](#string) |  |  |






<a name="bettor-v1alpha-GetBetResponse"></a>

### GetBetResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bet | [Bet](#bettor-v1alpha-Bet) |  |  |






<a name="bettor-v1alpha-GetMarketRequest"></a>

### GetMarketRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="bettor-v1alpha-GetMarketResponse"></a>

### GetMarketResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-GetUserByUsernameRequest"></a>

### GetUserByUsernameRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| book | [string](#string) |  |  |
| username | [string](#string) |  |  |






<a name="bettor-v1alpha-GetUserByUsernameResponse"></a>

### GetUserByUsernameResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#bettor-v1alpha-User) |  |  |






<a name="bettor-v1alpha-GetUserRequest"></a>

### GetUserRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="bettor-v1alpha-GetUserResponse"></a>

### GetUserResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| user | [User](#bettor-v1alpha-User) |  |  |






<a name="bettor-v1alpha-ListBetsRequest"></a>

### ListBetsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |
| book | [string](#string) |  |  |
| user | [string](#string) |  |  |
| market | [string](#string) |  |  |
| exclude_settled | [bool](#bool) |  |  |






<a name="bettor-v1alpha-ListBetsResponse"></a>

### ListBetsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bets | [Bet](#bettor-v1alpha-Bet) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="bettor-v1alpha-ListMarketsRequest"></a>

### ListMarketsRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |
| book | [string](#string) |  |  |
| status | [Market.Status](#bettor-v1alpha-Market-Status) |  |  |






<a name="bettor-v1alpha-ListMarketsResponse"></a>

### ListMarketsResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| markets | [Market](#bettor-v1alpha-Market) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="bettor-v1alpha-ListUsersRequest"></a>

### ListUsersRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| page_size | [int32](#int32) |  |  |
| page_token | [string](#string) |  |  |
| book | [string](#string) |  |  |
| users | [string](#string) | repeated |  |






<a name="bettor-v1alpha-ListUsersResponse"></a>

### ListUsersResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| users | [User](#bettor-v1alpha-User) | repeated |  |
| next_page_token | [string](#string) |  |  |






<a name="bettor-v1alpha-LockMarketRequest"></a>

### LockMarketRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name="bettor-v1alpha-LockMarketResponse"></a>

### LockMarketResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-Market"></a>

### Market
A betting market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| settled_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| title | [string](#string) |  |  |
| creator | [string](#string) |  |  |
| status | [Market.Status](#bettor-v1alpha-Market-Status) |  |  |
| pool | [Pool](#bettor-v1alpha-Pool) |  |  |






<a name="bettor-v1alpha-Outcome"></a>

### Outcome
An outcome in a pool betting market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| title | [string](#string) |  |  |
| centipoints | [uint64](#uint64) |  |  |






<a name="bettor-v1alpha-Pool"></a>

### Pool
Pool or parimutuel betting market.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| outcomes | [Outcome](#bettor-v1alpha-Outcome) | repeated |  |
| winner | [string](#string) |  |  |






<a name="bettor-v1alpha-SettleMarketRequest"></a>

### SettleMarketRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| winner | [string](#string) |  |  |






<a name="bettor-v1alpha-SettleMarketResponse"></a>

### SettleMarketResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| market | [Market](#bettor-v1alpha-Market) |  |  |






<a name="bettor-v1alpha-User"></a>

### User
User information.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| created_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| updated_at | [google.protobuf.Timestamp](#google-protobuf-Timestamp) |  |  |
| username | [string](#string) |  |  |
| centipoints | [uint64](#uint64) |  |  |





 


<a name="bettor-v1alpha-Market-Status"></a>

### Market.Status


| Name | Number | Description |
| ---- | ------ | ----------- |
| STATUS_UNSPECIFIED | 0 |  |
| STATUS_OPEN | 1 |  |
| STATUS_BETS_LOCKED | 2 |  |
| STATUS_SETTLED | 3 |  |
| STATUS_CANCELED | 4 |  |


 

 


<a name="bettor-v1alpha-BettorService"></a>

### BettorService
BettorService is a service for bets and predictions.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| CreateUser | [CreateUserRequest](#bettor-v1alpha-CreateUserRequest) | [CreateUserResponse](#bettor-v1alpha-CreateUserResponse) | CreateUser creates a new user. |
| GetUser | [GetUserRequest](#bettor-v1alpha-GetUserRequest) | [GetUserResponse](#bettor-v1alpha-GetUserResponse) | GetUser returns a user by name. |
| GetUserByUsername | [GetUserByUsernameRequest](#bettor-v1alpha-GetUserByUsernameRequest) | [GetUserByUsernameResponse](#bettor-v1alpha-GetUserByUsernameResponse) | GetUserByUsername returns a user by name. |
| ListUsers | [ListUsersRequest](#bettor-v1alpha-ListUsersRequest) | [ListUsersResponse](#bettor-v1alpha-ListUsersResponse) | ListUsers lists users by filters. |
| CreateMarket | [CreateMarketRequest](#bettor-v1alpha-CreateMarketRequest) | [CreateMarketResponse](#bettor-v1alpha-CreateMarketResponse) | CreateMarket creates a new betting market. |
| GetMarket | [GetMarketRequest](#bettor-v1alpha-GetMarketRequest) | [GetMarketResponse](#bettor-v1alpha-GetMarketResponse) | GetMarket gets a betting market by name. |
| ListMarkets | [ListMarketsRequest](#bettor-v1alpha-ListMarketsRequest) | [ListMarketsResponse](#bettor-v1alpha-ListMarketsResponse) | ListMarkets lists markets by filters. |
| LockMarket | [LockMarketRequest](#bettor-v1alpha-LockMarketRequest) | [LockMarketResponse](#bettor-v1alpha-LockMarketResponse) | LockMarket locks a betting market preventing further bets. |
| SettleMarket | [SettleMarketRequest](#bettor-v1alpha-SettleMarketRequest) | [SettleMarketResponse](#bettor-v1alpha-SettleMarketResponse) | SettleMarket settles a betting market and pays out bets. |
| CancelMarket | [CancelMarketRequest](#bettor-v1alpha-CancelMarketRequest) | [CancelMarketResponse](#bettor-v1alpha-CancelMarketResponse) | CancelMarket cancels a betting market and redunds all bettors. |
| CreateBet | [CreateBetRequest](#bettor-v1alpha-CreateBetRequest) | [CreateBetResponse](#bettor-v1alpha-CreateBetResponse) | CreateBet places a bet on an open betting market. |
| GetBet | [GetBetRequest](#bettor-v1alpha-GetBetRequest) | [GetBetResponse](#bettor-v1alpha-GetBetResponse) | GetBet gets a bet. |
| ListBets | [ListBetsRequest](#bettor-v1alpha-ListBetsRequest) | [ListBetsResponse](#bettor-v1alpha-ListBetsResponse) | ListBet lists bets by filters. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

