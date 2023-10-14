package configuration

type Config struct {
	System struct {
		Http struct {
			Port    int
			Timeout int
		}
		Postgres struct {
			Ip       string
			Port     int
			DbName   string
			User     string
			Password string
			Timeout  int
		}
		Kubernetes struct {
			Timeout int
		}
	}
	Logger struct {
		LogLevel string
	}
	ScanDelay      int
	JobGrepPattern string
}
