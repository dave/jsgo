package cmdconfig

type Config struct {
	Port     int
	Index    string
	Template string
	Json     bool
	Quiet    bool
	Open     bool
	Command  string
	Flags    string
	Path     string
}
