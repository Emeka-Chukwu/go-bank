package gapi

import (
	db "simplebank/db/sqlc"
	"simplebank/util"
	"simplebank/worker"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)
	return server
}
