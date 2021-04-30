package config

// DirectoryConfig represents the config of a directory that must
// be backed up
type DirectoryConfig struct {
	Local     string
	Remote    string
	Recursive bool
}

// Config holds all the config
type Config struct {
	Directories []DirectoryConfig
	TokenPath   string
}
