package etcd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/0xa1-red/empires-of-avalon/config"
	"github.com/0xa1-red/empires-of-avalon/pkg/service/blueprints"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
)

type Store struct {
	*clientv3.Client
}

func namespace(kind, userKey string) string {
	return fmt.Sprintf("%s%s%s%s%s",
		viper.GetString(config.Registry_Etcd_Key_Root),
		viper.GetString(config.Registry_Etcd_Key_Separator),
		kind,
		viper.GetString(config.Registry_Etcd_Key_Separator),
		userKey,
	)
}

func New() (*Store, error) {
	etcdConf := clientv3.Config{ // nolint
		Endpoints:   viper.GetStringSlice(config.ETCD_Endpoints),
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
		Username:    viper.GetString(config.ETCD_User),
		Password:    viper.GetString(config.ETCD_Passwd),
	}

	slog.Debug("creating etcd provider",
		"endpoints", viper.GetStringSlice(config.ETCD_Endpoints),
		"username", viper.GetString(config.ETCD_User),
	)

	c, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	return &Store{c}, nil
}

func (s *Store) Push(blueprint blueprints.Blueprint) error {
	buf := bytes.NewBufferString("")
	encoder := json.NewEncoder(buf)

	if err := encoder.Encode(blueprint); err != nil {
		return err
	}

	if _, err := s.Put(context.Background(), namespace(blueprint.Kind(), blueprint.GetName()), buf.String()); err != nil {
		return err
	}

	return nil
}

func (s *Store) List() (map[string]blueprints.Blueprint, error) {
	return nil, nil
}

func (s *Store) Get(kind, name string) (blueprints.Blueprint, error) {
	key := namespace(kind, name)

	slog.Debug("looking up blueprint", "kind", kind, "name", name, "key", key)

	resp, err := s.Client.Get(context.Background(), namespace(kind, name))
	if err != nil {
		return nil, err
	}

	if resp.Count == 0 {
		return nil, fmt.Errorf("Not found")
	}

	buf := bytes.NewBuffer(resp.Kvs[0].Value)
	decoder := json.NewDecoder(buf)

	var res blueprints.Blueprint

	switch kind {
	case blueprints.KindBuilding:
		res = &blueprints.Building{} // nolint:exhaustruct
	case blueprints.KindResource:
		res = &blueprints.Resource{} // nolint:exhaustruct
	}

	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	return res, nil
}
