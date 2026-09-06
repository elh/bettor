package server_test

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	otelconnect "github.com/bufbuild/connect-opentelemetry-go"
	api "github.com/elh/bettor/api/bettor/v1alpha"
	"github.com/elh/bettor/api/bettor/v1alpha/bettorv1alphaconnect"
	"github.com/elh/bettor/internal/app/bettor/repo/mem"
	"github.com/elh/bettor/internal/app/bettor/server"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/proto"
)

// TestTransportCompatibility exercises the generated clients and handlers over
// real connections, including the h2c and telemetry wrappers used in production.
func TestTransportCompatibility(t *testing.T) {
	testCases := []struct {
		name    string
		http2   bool
		options []connect.ClientOption
	}{
		{name: "connect protobuf"},
		{name: "connect JSON", options: []connect.ClientOption{connect.WithProtoJSON()}},
		{name: "gRPC h2c", http2: true, options: []connect.ClientOption{connect.WithGRPC()}},
		{name: "gRPC-Web", options: []connect.ClientOption{connect.WithGRPCWeb()}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := server.New(server.WithRepo(&mem.Repo{}))
			require.NoError(t, err)
			path, handler := bettorv1alphaconnect.NewBettorServiceHandler(s, otelconnect.WithTelemetry())
			mux := http.NewServeMux()
			mux.Handle(path, handler)
			ts := httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))
			t.Cleanup(ts.Close)
			httpClient := ts.Client()
			if tc.http2 {
				transport := &http2.Transport{
					AllowHTTP: true,
					DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
						return (&net.Dialer{}).DialContext(ctx, network, addr)
					},
				}
				t.Cleanup(transport.CloseIdleConnections)
				httpClient = &http.Client{Transport: transport}
			}
			client := bettorv1alphaconnect.NewBettorServiceClient(httpClient, ts.URL, tc.options...)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			created, err := client.CreateUser(ctx, connect.NewRequest(&api.CreateUserRequest{
				Book: "books/transport",
				User: &api.User{Username: "bettor", Centipoints: 123456},
			}))
			require.NoError(t, err)
			user := created.Msg.GetUser()
			require.NoError(t, user.Validate())
			require.Equal(t, uint64(123456), user.GetCentipoints())
			got, err := client.GetUser(ctx, connect.NewRequest(&api.GetUserRequest{Name: user.GetName()}))
			require.NoError(t, err)
			require.True(t, proto.Equal(user, got.Msg.GetUser()))

			market, err := client.CreateMarket(ctx, connect.NewRequest(&api.CreateMarketRequest{
				Book: "books/transport",
				Market: &api.Market{
					Title: "Transport compatibility", Creator: user.GetName(),
					Type: &api.Market_Pool{Pool: &api.Pool{Outcomes: []*api.Outcome{{Title: "yes"}, {Title: "no"}}}},
				},
			}))
			require.NoError(t, err)
			require.NoError(t, market.Msg.GetMarket().Validate())
			require.Len(t, market.Msg.GetMarket().GetPool().GetOutcomes(), 2)

			_, err = client.GetUser(ctx, connect.NewRequest(&api.GetUserRequest{}))
			require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}
