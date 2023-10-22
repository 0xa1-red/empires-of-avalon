package timer

import (
	"fmt"

	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (g *Grain) Encode() ([]byte, error) {
	if g.timer == nil {
		return nil, fmt.Errorf("timer not initialized")
	}

	t, err := timerPb(g.timer)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(t)
}

func timerPb(timer *Timer) (*protobuf.TimerSnapshot, error) {
	dataPb := &structpb.Struct{}

	if timer.Data != nil {
		var err error
		dataPb, err = structpb.NewStruct(timer.Data)

		if err != nil {
			return nil, err
		}
	}

	t := &protobuf.TimerSnapshot{
		ID:           timer.TimerID,
		Kind:         timer.Kind,
		InventoryID:  timer.InventoryID,
		ReplySubject: timer.Reply,
		Amount:       timer.Amount,
		Start:        timestamppb.New(timer.Start),
		Interval:     timer.Interval.String(),
		Data:         dataPb,
	}

	return t, nil
}
