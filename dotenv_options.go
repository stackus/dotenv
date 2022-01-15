package dotenv

type LoadOption interface {
	loadOption(c *envCfg) error
}

type ParseOption interface {
	parseOption(c *envCfg) error
}

type FilesOpt []string

// Files option to set the file names to read environment variables from
func Files(files ...string) FilesOpt {
	return files
}

// EnvironmentFiles option to set environment and local environment variable files
//
// .env.<environment>.local
// .env.local
// .env.<environment>
// .env
func EnvironmentFiles(environment string) FilesOpt {
	// exclude the .env.local file when env="test" to keep
	// in step with the original bkeeper/dotenv
	if environment == "test" {
		return []string{
			".env." + environment + ".local",
			".env." + environment,
			".env",
		}
	}

	return []string{
		".env." + environment + ".local",
		".env.local",
		".env." + environment,
		".env",
	}
}

func (o FilesOpt) loadOption(c *envCfg) error {
	c.files = o

	return nil
}

func (o FilesOpt) parseOption(c *envCfg) error {
	c.files = o

	return nil
}

type PathsOpt []string

// Paths option to set the paths to search for files in
func Paths(paths ...string) PathsOpt {
	return paths
}

func (o PathsOpt) loadOption(c *envCfg) error {
	c.paths = o

	return nil
}

func (o PathsOpt) parseOption(c *envCfg) error {
	c.paths = o

	return nil
}

type OverloadOpt bool

// Overload option to replace any ENV values with the values read from files
func Overload() OverloadOpt {
	return true
}

func (o OverloadOpt) loadOption(c *envCfg) error {
	c.overload = bool(o)

	return nil
}

type RequiredKeysOpt []string

// RequiredKeys option will perform a check for any missing keys after loading the files
func RequiredKeys(keys ...string) RequiredKeysOpt {
	return keys
}

func (o RequiredKeysOpt) loadOption(c *envCfg) error {
	c.requiredKeys = o

	return nil
}

type AllFilesRequiredOpt bool

// AllFilesRequired option is used to raise an error if any files are missing
func AllFilesRequired() AllFilesRequiredOpt {
	return true
}

func (AllFilesRequiredOpt) loadOption(c *envCfg) error {
	c.requireFiles = true

	return nil
}

func (AllFilesRequiredOpt) parseOption(c *envCfg) error {
	c.requireFiles = true

	return nil
}
