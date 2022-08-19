package indexer

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/pokt-foundation/pocket-go/provider"
	"github.com/pokt-foundation/utils-go/mock-client"
	testMock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIndexer_IndexBlockApps(t *testing.T) {
	c := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	reqProvider := provider.NewProvider("https://dummy.com", []string{})

	writerMock := &writerMock{}

	indexer := NewIndexer(reqProvider, writerMock)

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.QueryAppsRoute),
		http.StatusInternalServerError, "samples/query_apps.json")

	addresses, err := indexer.IndexBlockApps(30363)
	c.Equal(provider.Err5xxOnConnection, err)
	c.Empty(addresses)

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.QueryAppsRoute),
		http.StatusOK, "samples/query_apps_empty.json")

	addresses, err = indexer.IndexBlockApps(30363)
	c.Equal(ErrNoAppsToIndex, err)
	c.Empty(addresses)

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.QueryAppsRoute),
		http.StatusOK, "samples/query_apps.json")

	writerMock.On("WriteApps", testMock.Anything).Return(errors.New("forced failure")).Once()

	addresses, err = indexer.IndexBlockApps(30363)
	c.EqualError(err, "forced failure")
	c.Len(addresses, 1)
	c.Equal("98a18a38aa6826a55dccce19f607e3171cf14366", addresses[0])

	writerMock.On("WriteApps", testMock.Anything).Return(nil).Once()

	addresses, err = indexer.IndexBlockApps(30363)
	c.NoError(err)
	c.Len(addresses, 1)
	c.Equal("98a18a38aa6826a55dccce19f607e3171cf14366", addresses[0])
}
