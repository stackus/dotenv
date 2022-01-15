package dotenv

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	varsRe         = regexp.MustCompile(`(?m)(?:^|\A)\s*(?:export\s+)?([\w.]+)(?:\s*=\s*?|:\s+?)(\s*'(?:\\'|[^'])*'|\s*"(?:\\"|[^"])*"|(?:[^\s\r\n]|[ \t]+\w)+)?\s*?(?:#.*)?(?:$|\z)`)
	exportsRe      = regexp.MustCompile(`(?m)(?:^|\A)\s*export\s+([\w.]+)\s*(?:#.*)?(?:$|\z)`)
	quotesRe       = regexp.MustCompile(`(?m)(?:^|\A)(?:'(?:\\'|[^'])*|"(?:\\"|[^"])*|(?:[^\s\r\n]|[ \t]+\w)+)?(["'])?(?:$|\z)`)
	unescapeRe     = regexp.MustCompile(`\\([^$])`)
	substitutionRe = regexp.MustCompile(`(?m)(\\)?\${?(\w+)?}?`)
)

type envCfg struct {
	files        []string
	paths        []string
	overload     bool
	requiredKeys []string
	requireFiles bool
}

type envVars map[string]string

func Load(options ...LoadOption) error {
	cfg := &envCfg{
		files:        []string{".env"},
		paths:        []string{"."},
		overload:     false,
		requiredKeys: []string{},
		requireFiles: false,
	}

	for _, option := range options {
		err := option.loadOption(cfg)
		if err != nil {
			return err
		}
	}

	return load(cfg)
}

func Parse(options ...ParseOption) (map[string]string, error) {
	cfg := &envCfg{
		files:        []string{".env"},
		paths:        []string{"."},
		overload:     false,
		requiredKeys: []string{},
		requireFiles: false,
	}

	for _, option := range options {
		err := option.parseOption(cfg)
		if err != nil {
			return nil, err
		}
	}

	return parse(cfg)
}

func load(cfg *envCfg) error {
	files, err := buildFileList(cfg)
	if err != nil {
		return err
	}

	for _, file := range files {
		fileEnvs, err := parseFile(file, cfg.overload, cfg.requireFiles)
		if err != nil {
			return err
		}

		err = applyEnvs(fileEnvs, cfg.overload)
		if err != nil {
			return err
		}
	}

	err = checkRequiredKeys(cfg)
	if err != nil {
		return err
	}

	return nil
}

func parse(cfg *envCfg) (envVars, error) {
	parsedEnvs := make(envVars)

	files, err := buildFileList(cfg)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileEnvs, err := parseFile(file, cfg.overload, cfg.requireFiles)
		if err != nil {
			return nil, err
		}

		currentEnv := mergeEnvs(parsedEnvs, systemEnvs())
		appliedEnvs := make(envVars)

		for key, value := range fileEnvs {
			if currentValue, exists := currentEnv[key]; !exists {
				appliedEnvs[key] = value
			} else {
				appliedEnvs[key] = currentValue
			}
		}

		parsedEnvs = mergeEnvs(parsedEnvs, appliedEnvs)
	}

	return parsedEnvs, nil
}

func checkRequiredKeys(cfg *envCfg) error {
	currentEnv := systemEnvs()

	if len(cfg.requiredKeys) > 0 {
		missingKeys := make([]string, 0)
		for _, key := range cfg.requiredKeys {
			if _, exists := currentEnv[key]; !exists {
				missingKeys = append(missingKeys, key)
			}
		}

		if len(missingKeys) > 0 {
			return fmt.Errorf("missing required configuration key(s): %s", strings.Join(missingKeys, ", "))
		}
	}
	return nil
}

func applyEnvs(envs envVars, overload bool) error {
	currentEnv := systemEnvs()

	for key, value := range envs {
		if _, exists := currentEnv[key]; !exists || overload {
			err := os.Setenv(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildFileList(cfg *envCfg) ([]string, error) {
	envFiles := make([]string, 0)

	for _, path := range cfg.paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		info, statErr := os.Stat(absPath)
		if !(!errors.Is(statErr, fs.ErrNotExist) && info.IsDir()) {
			return nil, fmt.Errorf("path does not exist or is not a directory: %s", path)
		}

		for _, envFile := range cfg.files {
			envFiles = append(envFiles, filepath.Join(absPath, envFile))
		}
	}

	return envFiles, nil
}

func parseFile(fileName string, overload, mustExist bool) (envVars, error) {
	if info, err := os.Stat(fileName); errors.Is(err, fs.ErrNotExist) || info.IsDir() {
		if errors.Is(err, fs.ErrNotExist) && mustExist {
			return nil, fmt.Errorf("environment variables file was not found: %s", fileName)
		}
		return envVars{}, nil
	}

	contents, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return parseString(string(contents), overload)
}

func parseString(contents string, overload bool) (envVars, error) {
	matches := varsRe.FindAllStringSubmatch(contents, -1)

	parsedEnvs := make(envVars)

	for _, match := range matches {
		parsedEnvs[match[1]] = parseValue(match[2], combineEnvs(parsedEnvs, overload))
	}

	exports := exportsRe.FindAllStringSubmatch(varsRe.ReplaceAllString(contents, ""), -1)
	for _, export := range exports {
		if export[1] != "" {
			if _, exists := parsedEnvs[export[1]]; !exists {
				return parsedEnvs, fmt.Errorf("line %s has an unset variable", export[0])
			}
		}
	}

	return parsedEnvs, nil
}

func parseValue(value string, envs envVars) string {
	value = strings.Trim(value, " \t\f")
	m := quotesRe.FindStringSubmatch(value)
	quote := m[1]
	value = strings.Trim(value, quote)
	switch quote {
	case `"`:
		value = strings.ReplaceAll(strings.ReplaceAll(value, `\n`, "\n"), `\r`, "\r")
		fallthrough
	case ``:
		value = unescapeRe.ReplaceAllString(value, "$1")
	}

	if quote != "'" {
		value = substitutionRe.ReplaceAllStringFunc(value, func(s string) string {
			submatch := substitutionRe.FindStringSubmatch(s)

			if submatch[1] != "" {
				return submatch[0][1:]
			}

			if submatch[2] == "" {
				return s
			}

			if val, exists := envs[submatch[2]]; exists {
				return val
			}

			return ""
		})
	}

	return value
}

func systemEnvs() envVars {
	currentEnv := make(envVars)

	for _, line := range os.Environ() {
		pair := strings.SplitN(line, "=", 2)
		currentEnv[pair[0]] = pair[1]
	}

	return currentEnv
}

func combineEnvs(parsedEnvs envVars, overload bool) envVars {
	if overload {
		return mergeEnvs(systemEnvs(), parsedEnvs)
	}

	return mergeEnvs(parsedEnvs, systemEnvs())
}

func mergeEnvs(envs ...envVars) envVars {
	mergedEnvs := make(envVars)

	for _, env := range envs {
		for key, value := range env {
			mergedEnvs[key] = value
		}
	}

	return mergedEnvs
}
