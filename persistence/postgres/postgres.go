package postgres

import (
	"fmt"
	"strings"
	"time"

	"github.com/0xa1-red/empires-of-avalon/database"
	"github.com/0xa1-red/empires-of-avalon/persistence/contract"
	"github.com/0xa1-red/empires-of-avalon/protobuf"
	"github.com/asynkron/protoactor-go/cluster"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type Snapshot struct {
	Kind      string    `db:"kind"`
	Identity  uuid.UUID `db:"identity"`
	Data      []byte    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}

type restorableGrain interface {
	Restore(r *protobuf.RestoreRequest, opts ...cluster.GrainCallOption) (*protobuf.RestoreResponse, error)
}

type Persister struct {
	db *database.Conn
	c  *cluster.Cluster
}

func NewPersister(c *cluster.Cluster) *Persister {
	p := &Persister{
		c: c,
	}
	p.db = database.Connection()

	return p
}

func (p *Persister) Persist(item contract.Persistable) (int, error) {
	raw, err := item.Encode()
	if err != nil {
		return 0, err
	}

	if raw == nil {
		return 0, nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec("INSERT INTO snapshots (kind, identity, data) VALUES ($1, $2, $3)",
		item.Kind(),
		item.Identity(),
		raw,
	); err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return len(raw), nil
}

func (p *Persister) Restore(kind, identity string) error {
	query, params := buildRestoreQuery(kind, identity)

	slog.Debug("Attempting to restore actors", "kind", kind, "query", query)

	res := []Snapshot{}
	err := p.db.Select(&res, query, params...)
	if err != nil {
		return err
	}

	if len(res) == 0 {
		slog.Debug("no snapshot found", "kind", kind, "query", query)
		return nil
	}

	for _, item := range res {
		p.restore(item)
	}

	return nil
}

func (p *Persister) restore(item Snapshot) error {
	var client restorableGrain
	switch item.Kind {
	case "inventory":
		client = protobuf.GetInventoryGrainClient(p.c, item.Identity.String())
	case "timer":
		client = protobuf.GetTimerGrainClient(p.c, item.Identity.String())
	}

	res, _ := client.Restore(&protobuf.RestoreRequest{Data: item.Data})
	if res.Status == protobuf.Status_Error {
		return fmt.Errorf("%s", res.Error)
	}

	return nil
}

func buildRestoreQuery(kind, identity string) (string, []interface{}) {
	filter := make([]string, 0)
	params := make([]interface{}, 0)
	if kind != "" {
		filter = append(filter, "kind")
		params = append(params, kind)
	}

	if identity != "" {
		filter = append(filter, "identity")
		params = append(params, identity)
	}

	for i := range filter {
		filter[i] = fmt.Sprintf("%s = $%d", filter[i], i+1)
	}

	query := "SELECT DISTINCT ON (kind, identity) kind, identity, data, created_at FROM snapshots"
	if len(filter) > 0 {
		filterStr := strings.Join(filter, " AND ")
		query = fmt.Sprintf("%s WHERE %s", query, filterStr)
	}
	query = fmt.Sprintf("%s ORDER BY kind, identity, created_at DESC", query)

	return query, params
}
