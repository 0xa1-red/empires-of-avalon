package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	if err := setupRegistry(); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	g := &Grain{}

	var assetError error

	g.buildings, assetError = g.getStartingBuildings()
	assert.NoError(t, assetError)

	g.resources, assetError = g.getStartingResources()
	assert.NoError(t, assetError)

	g.updateLimits()

	raw, err := g.Encode()
	assert.NoError(t, err)

	err = g.Restore(raw)
	assert.NoError(t, err)
}
