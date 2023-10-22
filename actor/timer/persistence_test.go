package timer

import (
	"testing"
	"time"

	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTimerPb(t *testing.T) {
	g := &Grain{}

	id := uuid.New().String()
	inventoryID := uuid.New().String()
	reply := "bogus"
	now := time.Now().UTC()

	g.timer = &Timer{
		TimerID:     id,
		Kind:        protobuf.TimerKind_Building,
		InventoryID: inventoryID,
		Reply:       reply,
		Amount:      1,
		Start:       now,
		Interval:    10 * time.Minute,
		Data: map[string]interface{}{
			"": nil,
		},
	}

	pb, err := timerPb(g.timer)
	assert.NoError(t, err)

	assert.Equal(t, id, pb.GetID())
	assert.Equal(t, inventoryID, pb.GetInventoryID())
	assert.Equal(t, protobuf.TimerKind_Building, pb.GetKind())
	assert.Equal(t, reply, pb.GetReplySubject())
	assert.Equal(t, now, pb.GetStart().AsTime())
	assert.Equal(t, int64(1), pb.GetAmount())
	assert.Equal(t, "10m0s", pb.GetInterval())
}
