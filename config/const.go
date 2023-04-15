package config

const (
	envPrefix = "AVALOND"
)

const (
	DB_Kind = "Database.Kind"
)

const (
	PG_User    = "Database.Postgres.User"
	PG_Passwd  = "Database.Postgres.Password"
	PG_Host    = "Database.Postgres.Host"
	PG_Port    = "Database.Postgres.Port"
	PG_DB      = "Database.Postgres.Database"
	PG_SSLMode = "Database.Postgres.SSLMode"
)

const (
	Cluster_Name = "Cluster.Name"
	Node_Host    = "Cluster.Node.Host"
	Node_Port    = "Cluster.Node.Port"
)

const (
	HTTP_Address = "HTTP.Address"
	HTTP_Port    = "HTTP.Port"
)

const (
	ETCD_Endpoints = "ETCD.Endpoints"
	ETCD_Root      = "ETCD.Root"
)

const (
	NATS_Host     = "NATS.Host"
	NATS_Port     = "NATS.Port"
	NATS_User     = "NATS.User"
	NATS_Password = "NATS.Password"
)

const (
	Logging_Level = "Logging.Level"
)

const (
	Persistence_Encoding = "Persistence.Encoding"

	EncodingGob  = "gob"
	EncodingJson = "json"
)
