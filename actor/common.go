package actor

import (
	"github.com/0xa1-red/empires-of-avalon/actor/admin"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/0xa1-red/empires-of-avalon/transport/nats"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func SendUpdate(update *protobuf.GrainUpdate) error {
	conn, err := nats.GetConnection()
	if err != nil {
		slog.Error("failed to get NATS connection", err)
		return err
	}

	if err := conn.Publish(admin.AdminSubject, update); err != nil {
		return err
	}

	return nil
}

func Register(kind protobuf.GrainKind, identity string) error {
	return SendUpdate(&protobuf.GrainUpdate{
		UpdateKind: protobuf.UpdateKind_Register,
		GrainKind:  kind,
		Timestamp:  timestamppb.Now(),
		Identity:   identity,
	})
}

func Heartbeat(kind protobuf.GrainKind, identity string) error {
	return SendUpdate(&protobuf.GrainUpdate{
		UpdateKind: protobuf.UpdateKind_Heartbeat,
		GrainKind:  kind,
		Timestamp:  timestamppb.Now(),
		Identity:   identity,
	})
}
