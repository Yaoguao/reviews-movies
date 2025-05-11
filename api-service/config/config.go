package config

type Config struct {
	Port    int
	Env     string
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
}
