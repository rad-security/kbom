package kube

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListPages(t *testing.T) {
	pages := [][]string{{"a", "b"}, {"c"}, {"d"}}

	var calls []metav1.ListOptions
	items, err := listPages(context.Background(), func(opts metav1.ListOptions) ([]string, string, error) {
		calls = append(calls, opts)

		page := len(calls) - 1
		next := ""
		if page < len(pages)-1 {
			next = strconv.Itoa(page + 1)
		}
		return pages[page], next, nil
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c", "d"}, items)
	require.Len(t, calls, 3)
	assert.Equal(t, int64(listPageSize), calls[0].Limit)
	assert.Empty(t, calls[0].Continue)
	assert.Equal(t, "1", calls[1].Continue)
	assert.Equal(t, "2", calls[2].Continue)
}

func TestListPagesError(t *testing.T) {
	_, err := listPages(context.Background(), func(metav1.ListOptions) ([]string, string, error) {
		return nil, "", errors.New("forbidden")
	})
	assert.EqualError(t, err, "forbidden")
}

func TestListPagesCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := listPages(ctx, func(metav1.ListOptions) ([]string, string, error) {
		t.Fatal("list must not be called with a canceled context")
		return nil, "", nil
	})
	assert.ErrorIs(t, err, context.Canceled)
}
