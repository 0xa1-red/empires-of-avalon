package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPIDFromIdentity(t *testing.T) {
	p, err := PIDFromIdentity("Address:\"127.0.0.1:52479\" Id:\"partition-activator/6751d512-e594-5f0b-b470-c2152ccb03ac$2P\"")

	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:52479", p.GetAddress())
	assert.Equal(t, "partition-activator/6751d512-e594-5f0b-b470-c2152ccb03ac$2P", p.GetId())
	assert.Equal(t, "6751d512-e594-5f0b-b470-c2152ccb03ac", p.GetGrainID().String())
}
