package config

const (
	envPrefix = "AVALOND"
)

const (
	PG_User    = "Postgres.User"
	PG_Passwd  = "Postgres.Password"
	PG_Host    = "Postgres.Host"
	PG_Port    = "Postgres.Port"
	PG_DB      = "Postgres.Database"
	PG_SSLMode = "Postgres.SSLMode"
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
