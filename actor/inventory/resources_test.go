package inventory

import (
	"testing"

	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCap(t *testing.T) {
	if err := setupRegistry(); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	g := &Grain{}
	g.buildings = g.getStartingBuildings()
	g.resources = g.getStartingResources()

	resource := g.resources[blueprints.Wood]

	assert.Equal(t, 0, resource.Cap)

	if err := resource.UpdateCap(g.resources, g.buildings); err != nil {
		t.Fatalf("Fail: %v", err)
	}

	assert.Equal(t, 100, resource.Cap)
}
