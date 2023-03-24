package service

type MetricConfig struct {
	Addr string `json:"addr" env:"METRIC_ADDRESS"`
}

type ConsulConfig struct {
	Addr          string `json:"addr" env:"consul_host"`
	Token         string `json:"token" env:"consul_config_acl_token"`
	AppPath       string `json:"path" env:"consul_app_path"`
	NamespacePath string `json:"config_path" env:"consul_config_path"`
	WatchWait     string `json:"watch_wait"`
	WatchTimeout  string `json:"watch_timeout"`
}

type VaultConfig struct {
	Uri           string `json:"uri" env:"vault_uri"`
	NamespacePath string `json:"path" env:"vault_config_path"`
	AppPath       string `json:"apppath" env:"vault_app_path"`
	Token         string `json:"token" env:"vault_token"`
	SecretId      string `json:"secretid" env:"vault_secret_id"`
	RoleId        string `json:"roleid" env:"vault_role_id"`
}

type PostgresConfig struct {
	Addr            string `json:"addr"`
	Login           string `json:"login"`
	Passw           string `json:"passw"`
	DBName          string `json:"dbname"`
	AppName         string `json:"appname"`
	ConnMax         int    `json:"conn_max"`
	ConnMaxIdle     int    `json:"conn_max_idle"`
	ConnLifetime    int    `json:"conn_lifetime"`
	ConnMaxIdleTime int    `json:"conn_maxidletime"`
}

type OracleConfig struct {
	Addr            string `json:"addr"`
	Login           string `json:"login"`
	Passw           string `json:"passw"`
	DBName          string `json:"dbname"`
	ConnMax         int    `json:"conn_max"`
	ConnMaxIdle     int    `json:"conn_max_idle"`
	ConnLifetime    int    `json:"conn_lifetime"`
	ConnMaxIdleTime int    `json:"conn_maxidletime"`
}

type ServerConfig struct {
	Name    string `json:"name" env:"SERVER_NAME"`
	ID      string `json:"id" env:"SERVER_ID"`
	Version string `json:"version" env:"SERVER_VERSION"`
	Addr    string `json:"addr" env:"SERVER_ADDRESS"`
}

type CoreConfig struct {
	GetConfig bool `json:"-"`
	Profile   bool `json:"profile"`
}
