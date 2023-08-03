package admin

import (
	"strings"

	pactor "github.com/asynkron/protoactor-go/actor"
	"golang.org/x/exp/slog"
)

func PIDFromIdentity(identity string) *pactor.PID {
	parts := strings.Split(identity, " ")

	var (
		address string
		id      string
	)

	for _, part := range parts {
		kv := strings.Split(part, ":")
		switch kv[0] {
		case "Address":
			address = strings.Trim(strings.Join(kv[1:], ":"), `\"`)
		case "Id":
			id = strings.Trim(kv[1], `\"`)
		}
	}

	slog.Debug("parsed PID from identity", "address", address, "identity", id)

	return pactor.NewPID(address, id)
}
