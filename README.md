# dotenv

Yet another Go port of [dotenv](https://github.com/bkeepers/dotenv) environment variable dot-file library.

## Installation

```shell
go get github.com/stackus/dotenv
```

## Usage

Add your application configuration to your `.env` file in the root of your project:

```shell
S3_BUCKET=YOURS3BUCKET
SECRET_KEY=YOURSECRETKEYGOESHERE
```

### Load()

Use `Load()` to inject the variables read from the files into the current processes environment variables.

Load the variables in Go like the following:

```go
package main

import (
	"fmt"
	"os"
	
	"github.com/stackus/dotenv"
)

func main() {
    if err := dotenv.Load(); err != nil {
        fmt.Fprintf(os.Stderr, "error starting application: %s", err)
        os.Exit(1)
    }
	
    s3Bucket := os.Getenv("S3_BUCKET")
    secretKey := os.Getenv("SECRET_KEY")
	
    // ...
}
```

### Parse()

If you do not want to alter the environment variables you can use `Parse()` to return all of the values read from the file as a `map[string]string`.

```go
// ...

values, err := dotenv.Parse()

s3Bucket := values["S3_BUCKET"]

// ...
```

### Defaults
| Setting | Default | Purpose                                                               |
| --- | --- |-----------------------------------------------------------------------|
| Files  | .env | Names of the files that should be parsed for values                   |
| Paths | . | Paths that should be searched for files                               |
| Overload | false | Replace existing environment variables with values read in from files |
| RequireAllFiles | false | Silently skip any files that could not be found and read              |
| RequiredKeys | [] | List of keys that must exist in the environment                       |
### Options

Both `Load()` and `Parse()` accept options that will alter how they work.

#### Files(...string)
Provide a list of file names to read values from.

#### Paths(...string)
Provide a list of paths to search for files with values.

#### Overload()
> This is a `Load()` only option.

Replace any existing values that either were already set in the environment variables or were set from a previously read file.

#### RequiredKeys(...string)
> This is a `Load()` only option.

Provides a list of keys that will be checked just before `Load()` is done. If any of the keys are not set in the environment then `Load()` will return an error message listing all missing keys.

#### AllFilesRequired()
This will cause either `Load()` or `Parse()` to return an error when the first missing file is encountered.

#### EnvironmentFiles(string)
Sets a group of files using the given environment name.

| environment | Result                                                           |
| --- |------------------------------------------------------------------|
| test | .env.test.local, .env.test, .env                                 |
| anything else | .env.\<environment>.local, .env.local, .env.\<environment>, .env |

#### Using Options

You can pass in any combination of options you need to either `Load()` or `Parse()`.

```go

err := dotenv.Load(
        dotenv.EnvironmentFiles(os.Getenv("MY_APP_ENV")),
        dotenv.RequireKeys("DATABASE_URL", "HOST", "LOG_LEVEL"),
    )

// the same goes for Parse()
```

## Autoload
There is an autoload as well if you do not need to make use of any of the options.

```go
package main

import (
	_ "github.com/stackus/dotenv/autoload"
)

func main() {
    // values have already been loaded
}

```

## CLI
You can also use this library in the CLI to execute applications with modified environments.

### Installation

```shell
go install github.com/stackus/dotenv/cmd/dotenv@latest
```

### Usage
Prefix the command with a call to dotenv first along with some arguments to load the file or files you need, such as:

```shell
dotenv -f .env -f .other-env -- your-command -a your-args
```

In the above example `dotenv` was used to load values into the environment from two different files, then the command `your-command` and its arguments followed a split, identified as `--`. 

Use `--` to signal when the `dotenv` flag parser should stop. Everything to the right of it will be left unparsed and used as the command that will be run along with any arguments that may also exist.

The `dotenv` command will also accept the `-e` flag to set the environment which works like the `EnvironmentFiles(env)` option above, as well as the `-p` flag to provide one or more paths. `-p` may be repeated just like `-f`.

### Similarities with the Ruby version

Nearly everything the Ruby version would parse is parsed in this version. With one major difference. There is no command substitution.

```env
# start simple and go from there
FOO=bar             # bar
FOO1=bar#baz        # bar#baz
FOO2=fizz buzz      # fizz buzz
FOO3="bar"          # bar
FOO4= "bar"         # bar
FOO5 = "bar"        # bar
FOO6="foo$FOO5"     # foobar
FOO7="foo\$FOO5"    # foo$FOO5
FOO8="a multi
line value"         # a multi\nline value
BAR1='foo'          # foo
BAR2='bar$BAR1'     # bar$BAR1
BAR3: yaml-like     # yaml-like
export BAR4=bar     # bar
export UNDEFINED    # this will return a warning error just as the Ruby version does
```

> It also doesn't parse `"FOO=foo\rBAR=bar"` (strings that use only a carriage return) correctly.

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
