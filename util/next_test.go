package util

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNextFlow(t *testing.T) {
	t.Parallel()

	// Respect the default
	defaultURL := "/users/hai?aoo=bar"
	assert.Equal(t, defaultURL, NextFlow(defaultURL, url.Values{}))

	// Respect default values
	assert.Equal(t, defaultURL+"&stuff=yeah", NextFlow(defaultURL+"&stuff=yeah", url.Values{
		"next":  []string{"/users/hai"},
		"stuff": []string{"nah"},
	}))

	// Respect root default
	assert.Equal(t, "", NextFlow("", url.Values{}))
}
