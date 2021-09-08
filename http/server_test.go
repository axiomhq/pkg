package http_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	xhttp "github.com/axiomhq/pkg/http"
)

func TestServer(t *testing.T) {
	// A slow handler function.
	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 2)
		_, _ = w.Write([]byte("hello world"))
	})

	srv, err := xhttp.NewServer("localhost:0", hf)
	require.NoError(t, err)
	require.NotNil(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv.Run(ctx)

	go func() {
		err = <-srv.ListenError()
		require.NoError(t, err)
	}()

	req, err := http.NewRequest(http.MethodGet, "http://"+srv.ListenAddr().String(), nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var wasCancelled int32
	go func() {
		time.Sleep(time.Second)
		cancel()
		atomic.StoreInt32(&wasCancelled, 1)
	}()

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.EqualValues(t, 1, atomic.LoadInt32(&wasCancelled))
}
