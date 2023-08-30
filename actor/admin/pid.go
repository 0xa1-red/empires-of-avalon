package admin

import (
	"fmt"
	"regexp"
	"strings"

	pactor "github.com/asynkron/protoactor-go/actor"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

func PIDFromIdentity(identity string) (*ActorPID, error) {
	pattern, err := regexp.Compile(`Address:"([^"]+)" +Id:"([^/]+)/([a-z0-9\-]+)([^"]+)"`) // nolint:gocritic
	if err != nil {
		return nil, err
	}

	matches := pattern.FindStringSubmatch(strings.ReplaceAll(`\"`, `"`, identity))

	address := matches[1]
	namespace := matches[2]
	idStr := matches[3]
	suffix := matches[4]
	shortId := fmt.Sprintf("%s/%s%s", namespace, idStr, suffix)
	pid := pactor.NewPID(address, shortId)

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	slog.Debug("parsed PID from identity", "identity", identity, "address", address, "namespace", namespace, "id", id.String())

	return &ActorPID{
		PID:     pid,
		GrainID: id,
	}, nil
}
