package actor

import (
	"github.com/0xa1-red/empires-of-avalon/actor/admin"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/0xa1-red/empires-of-avalon/transport/nats"
	"golang.org/x/exp/slog"
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
