package inventory

import (
	"testing"

	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
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

	snapshot := &protobuf.InventorySnapshot{}
	err = proto.Unmarshal(raw, snapshot)
	assert.NoError(t, err)

	err = g.restore(snapshot)
	assert.NoError(t, err)
}
