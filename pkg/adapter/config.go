package adapter

type Configuration struct {
	ListenAddress string

	TLSCert string
	TLSKey  string

	CloudwatchNamespace string

	Debug bool
}
