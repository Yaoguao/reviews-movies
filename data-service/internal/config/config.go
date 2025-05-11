package config

type Config struct {
	Port int
	Env  string
	DB   struct {
		Dsn          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}
