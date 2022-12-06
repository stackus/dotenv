package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/stackus/dotenv"
)

type flagStrSlice []string

var files flagStrSlice
var paths flagStrSlice
var environment string

func main() {
	flag.Var(&files, "f", "[optional] [repeatable] files with key:value pairs to set into the current environment")
	flag.StringVar(&environment, "e", "", "[optional] sets the environment to load a suite of files")
	flag.Var(&paths, "p", "[optional] [repeatable] one or more paths to search for files")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(out, "\nExamples:")
		fmt.Fprintln(out, "Multiple files:\n\t dotenv -f .env -f .another.env -- some_command -a args")
		fmt.Fprintln(out, "Environment and paths:\n\t dotenv -e development -p ../devcfg -- some_command -a args")
	}

	flag.Parse()

	var options []dotenv.ParseOption

	// parse some files
	if len(files) > 0 {
		options = append(options, dotenv.Files(files...))
	}

	// parse a suite of files based on the provided environment
	if environment != "" {
		options = append(options, dotenv.EnvironmentFiles(environment))
	}

	// look for files in other paths
	if len(paths) > 0 {
		options = append(options, dotenv.Paths(paths...))
	}

	// parse everything into a map
	vars, err := dotenv.Parse(options...)
	if err != nil {
		log.Fatal("loading environment files errored: ", err)
	}

	if len(flag.Args()) == 0 {
		return
	}

	// turn the map into a slice of "k=v" strings
	var env []string
	for k, v := range vars {
		env = append(env, k+"="+v)
	}

	err = runCommand(flag.Args(), env)
	if err != nil {
		log.Fatal("encountered an error spawning command: ", err)
	}
}

func runCommand(args, env []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)
	// is Start() and Wait() what we want here?
	err = cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

// String implements flag.Value and fmt.Stringer to allow the value to be rendered as a plain string
func (v *flagStrSlice) String() string {
	return strings.Join(*v, " ")
}

// Set implements flag.Value so that multiple flag usages will build a list of strings vs setting a single one
func (v *flagStrSlice) Set(s string) error {
	*v = append(*v, s)

	return nil
}
