package config

var (
	RedisSentinel = map[string][]string{
		"production": {"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
		"preview":    {"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
		"test":       {"127.0.0.1:26379", "127.0.0.1:26380", "127.0.0.1:26381"},
	}

	RedisGroup = map[string][]string{
		"production": {"127.0.0.1:6379", "127.0.0.1:6380"},
		"preview":    {"127.0.0.1:6379", "127.0.0.1:6380"},
		"test":       {"127.0.0.1:6379", "127.0.0.1:6380"},
	}

	MysqlConf = map[string]map[string]string{
		"production": {
			"MYSQL_MUSIC_HOST": "127.0.0.1:3306",
			"MYSQL_USER":       "admin",
			"MYSQL_PASSWD":     "secret",
			"MYSQL_DB_NAME":    "test"},

		"preview": {
			"MYSQL_MUSIC_HOST": "127.0.0.1:3306",
			"MYSQL_USER":       "admin",
			"MYSQL_PASSWD":     "secret",
			"MYSQL_DB_NAME":    "test"},

		"test": {
			"MYSQL_MUSIC_HOST": "127.0.0.1:3306",
			"MYSQL_USER":       "admin",
			"MYSQL_PASSWD":     "secret",
			"MYSQL_DB_NAME":    "test"},
	}
)

const (
	ENV = "production"
	// redis config
	RedisSentinelName = "master"
	RedisSentinelPass = "secret"
	MYSQL_TPL         = "%s:%s@tcp(%s)/%s?charset=utf8"
)
