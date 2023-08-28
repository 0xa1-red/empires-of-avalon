// nolint:gosec
package config

const (
	envPrefix = "AD"
)

const (
	BlueprintPath = "blueprint_path"
)

const (
	PG_User    = "postgres.user"
	PG_Passwd  = "postgres.password"
	PG_Host    = "postgres.host"
	PG_Port    = "postgres.port"
	PG_DB      = "postgres.database"
	PG_SSLMode = "postgres.sslmode"
)

const (
	Cluster_Name = "cluster.name"
	Node_Host    = "cluster.node.host"
	Node_Port    = "cluster.node.port"
)

const (
	HTTP_Address = "http.address"
	HTTP_Port    = "http.port"
)

const (
	ETCD_Endpoints = "etcd.endpoints"
	ETCD_Root      = "etcd.root"
	ETCD_User      = "etcd.user"
	ETCD_Passwd    = "etcd.password"
)

const (
	NATS_Host     = "nats.host"
	NATS_Port     = "nats.port"
	NATS_User     = "nats.user"
	NATS_Password = "nats.password"
)

const (
	Logging_Level = "logging.level"
	Logging_Path  = "logging.path"
)

const (
	Persistence_Encoding = "persistence.encoding"

	EncodingGob  = "gob"
	EncodingJson = "json"
)

const (
	Instrumentation_Traces_Endpoint = "instrumentation.traces.endpoint"
	Instrumentation_Traces_Insecure = "instrumentation.traces.insecure"
)

const (
	Authenticator_Domain        = "authenticator.domain"
	Authenticator_Client_ID     = "authenticator.client_id"
	Authenticator_Client_Secret = "authenticator.client_secret"
	Authenticator_Audience      = "authenticator.audience"
)

const (
	Registry_Remote_Kind        = "registry.remote_kind"
	Registry_Etcd_Key_Root      = "registry.etcd.key_root"
	Registry_Etcd_Key_Separator = "registry.etcd.key_separator"
)
