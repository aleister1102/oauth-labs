package victim

type Credentials struct {
	Username string
	Password string
}

type LabConfig struct {
	LabNumber   string
	ServerURL   string
	Credentials Credentials
	ClientURL   string
}
