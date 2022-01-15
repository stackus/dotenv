package dotenv

import (
	"os"
	"reflect"
	"testing"
)

func TestParsing(t *testing.T) {
	type args struct {
		contents string
		overload bool
	}
	tests := map[string]struct {
		args    args
		setEnvs envVars
		want    envVars
		wantErr bool
	}{
		"parses unquoted values": {
			args:    args{"FOO=bar", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses unquoted values with spaces after separator": {
			args:    args{"FOO= bar", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses unquoted values with spaces before separator": {
			args:    args{"FOO =bar", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses unquoted escape characters correctly": {
			args:    args{"FOO=bar\\ bar", false},
			want:    envVars{"FOO": "bar bar"},
			wantErr: false,
		},
		"parses values with leading spaced": {
			args:    args{"  FOO=bar", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses values with following spaced": {
			args:    args{"FOO=bar  ", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses double quoted values": {
			args:    args{`FOO="bar"`, false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses double quoted values with following spaces": {
			args:    args{`FOO="bar"  `, false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses single quoted values": {
			args:    args{`FOO='bar'`, false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses single quoted values with following spaces": {
			args:    args{`FOO='bar'  `, false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"parses escaped double quotes": {
			args:    args{`FOO="escaped\"bar"`, false},
			want:    envVars{"FOO": `escaped"bar`},
			wantErr: false,
		},
		"parses empty values": {
			args:    args{`FOO=`, false},
			want:    envVars{"FOO": ``},
			wantErr: false,
		},
		"parses variables found in values": {
			args:    args{"FOO=test\nBAR=$FOO", false},
			want:    envVars{"FOO": `test`, "BAR": `test`},
			wantErr: false,
		},
		"parses variables found in values and ENV": {
			args:    args{"FOO=$FIZZ\nBAR=$FOO", false},
			setEnvs: envVars{"FIZZ": "test"},
			want:    envVars{"FOO": `test`, "BAR": `test`},
			wantErr: false,
		},
		"parses variables wrapped in brackets": {
			args:    args{"FOO=test\nBAR=bar${FOO}", false},
			want:    envVars{"FOO": `test`, "BAR": `bartest`},
			wantErr: false,
		},
		"expands variables from ENV if not found in .env": {
			args:    args{`BAR=$FOO`, false},
			setEnvs: envVars{"FOO": "test"},
			want:    envVars{"BAR": "test"},
			wantErr: false,
		},
		"expands variables from ENV if found in .env without overload": {
			args:    args{"FOO=development\nBAR=$FOO", false},
			setEnvs: envVars{"FOO": "test"},
			want:    envVars{"FOO": "development", "BAR": "test"},
			wantErr: false,
		},
		"expands variables from .env if found in ENV with overload": {
			args:    args{"FOO=development\nBAR=$FOO", true},
			setEnvs: envVars{"FOO": "test"},
			want:    envVars{"FOO": "development", "BAR": "development"},
			wantErr: false,
		},
		"expands undefined variables to and empty string": {
			args:    args{`BAR=$FOO`, false},
			want:    envVars{"BAR": ""},
			wantErr: false,
		},
		"expands variables in double quoted strings": {
			args:    args{"FOO=test\nBAR=\"$FOO\"", false},
			want:    envVars{"FOO": "test", "BAR": "test"},
			wantErr: false,
		},
		"does not expand variables in single quoted strings 1": {
			args:    args{"FOO=test\nBAR='$FOO'", false},
			want:    envVars{"FOO": "test", "BAR": "$FOO"},
			wantErr: false,
		},
		"does not expand variables in single quoted strings 2": {
			args:    args{"FOO=test\nBAR='${FOO}'", false},
			want:    envVars{"FOO": "test", "BAR": "${FOO}"},
			wantErr: false,
		},
		"does not expand escaped variables 1": {
			args:    args{"FOO=\"foo\\$BAR\"", false},
			want:    envVars{"FOO": "foo$BAR"},
			wantErr: false,
		},
		"does not expand escaped variables 2": {
			args:    args{"FOO=\"foo\\${BAR}\"", false},
			want:    envVars{"FOO": "foo${BAR}"},
			wantErr: false,
		},
		"does not expand escaped variables 3": {
			args:    args{"FOO=test\nBAR=\"foo\\${FOO} ${FOO}\"", false},
			want:    envVars{"FOO": "test", "BAR": "foo${FOO} test"},
			wantErr: false,
		},
		"parses yaml style options": {
			args:    args{`OPTION_A: 1`, false},
			want:    envVars{"OPTION_A": `1`},
			wantErr: false,
		},
		"parses export keyword": {
			args:    args{`export OPTION_A=2`, false},
			want:    envVars{"OPTION_A": `2`},
			wantErr: false,
		},
		"allows export line if you want to do it that way": {
			args:    args{"OPTION_A=2\nexport OPTION_A", false},
			want:    envVars{"OPTION_A": `2`},
			wantErr: false,
		},
		"allows export line if you want to do it that way and checks for unset variables": {
			args:    args{"OPTION_A=2\nexport OH_NO_NOT_SET", false},
			want:    envVars{"OPTION_A": `2`},
			wantErr: true,
		},
		"expands newlines in quoted strings": {
			args:    args{"FOO=\"bar\\nbaz\"", false},
			want:    envVars{"FOO": "bar\nbaz"},
			wantErr: false,
		},
		"parses variables with '.' in the name": {
			args:    args{"FOO.BAR=foobar", false},
			want:    envVars{"FOO.BAR": "foobar"},
			wantErr: false,
		},
		"strips unquoted values": {
			args:    args{"FOO=bar  ", false},
			want:    envVars{"FOO": "bar"},
			wantErr: false,
		},
		"ignores lines that are not variable assignments": {
			args:    args{"lol$wut", false},
			want:    envVars{},
			wantErr: false,
		},
		"ignores empty lines": {
			args:    args{"\n \t  \nfoo=bar\n \nfizz=buzz", false},
			want:    envVars{"foo": "bar", "fizz": "buzz"},
			wantErr: false,
		},
		"ignores inline comments": {
			args:    args{"foo=bar # this is foo", false},
			want:    envVars{"foo": "bar"},
			wantErr: false,
		},
		"allows '#' in double quoted value": {
			args:    args{"foo=\"bar#baz\" # comment\nbar=\"bar#baz\"\t#\tcomment", false},
			want:    envVars{"foo": "bar#baz", "bar": "bar#baz"},
			wantErr: false,
		},
		"allows '#' in single quoted value": {
			args:    args{"foo='bar#baz' # comment\nbar='bar#baz'\t#\tcomment", false},
			want:    envVars{"foo": "bar#baz", "bar": "bar#baz"},
			wantErr: false,
		},
		"allows '#' in unquoted value": {
			args:    args{"foo=bar#baz # comment\nbar=bar#baz\t#\tcomment", false},
			want:    envVars{"foo": "bar#baz", "bar": "bar#baz"},
			wantErr: false,
		},
		"ignores comment lines": {
			args:    args{"\n\n\n # HERE GOES FOO \nfoo=bar", false},
			want:    envVars{"foo": "bar"},
			wantErr: false,
		},
		"ignores commented out variables": {
			args:    args{"# HELLO=world\n", false},
			want:    envVars{},
			wantErr: false,
		},
		"ignores comment": {
			args:    args{"# Uncomment to activate:\n", false},
			want:    envVars{},
			wantErr: false,
		},
		"includes variables without values": {
			args: args{
				"DATABASE_PASSWORD=\nDATABASE_USERNAME=root\nDATABASE_HOST=/tmp/mysql.sock",
				false,
			},
			want:    envVars{"DATABASE_PASSWORD": "", "DATABASE_USERNAME": "root", "DATABASE_HOST": "/tmp/mysql.sock"},
			wantErr: false,
		},
		"allows multi-line values in single quotes": {
			args: args{
				`OPTION_A=first line
OPTION_B='line 1
line 2
line 3'
OPTION_C="last line"
OPTION_ESCAPED='line one
this is \\'quoted\\'
one more line'`,
				false,
			},
			want: envVars{
				"OPTION_A": "first line",
				"OPTION_B": "line 1\nline 2\nline 3",
				"OPTION_C": "last line",
				"OPTION_ESCAPED": `line one
this is \\'quoted\\'
one more line`,
			},
			wantErr: false,
		},
		"allows multi-line values in double quotes": {
			args: args{
				`OPTION_A=first line
OPTION_B="line 1
line 2
line 3"
OPTION_C="last line"
OPTION_ESCAPED="line one
this is \\"quoted\\"
one more line"`,
				false,
			},
			want: envVars{
				"OPTION_A": "first line",
				"OPTION_B": "line 1\nline 2\nline 3",
				"OPTION_C": "last line",
				"OPTION_ESCAPED": `line one
this is \"quoted\"
one more line`,
			},
			wantErr: false,
		},
		// TODO support carriage return by itself
		// "supports carriage return": {
		// 	args:    args{"FOO=bar\rbaz=fbb", false, nil},
		// 	want:    envVars{"FOO": "bar", "baz": "fbb"},
		// 	wantErr: false,
		// },
		"supports carriage return combined with new line": {
			args:    args{"FOO=bar\r\nbaz=fbb", false},
			want:    envVars{"FOO": "bar", "baz": "fbb"},
			wantErr: false,
		},
		"expands carriage return in quoted strings": {
			args:    args{"FOO=\"bar\\rbaz\"", false},
			want:    envVars{"FOO": "bar\rbaz"},
			wantErr: false,
		},
		"escape $ properly when no alphabets/numbers/_  are followed by it": {
			args:    args{`FOO="bar\$ \$\$"`, false},
			want:    envVars{"FOO": "bar$ $$"},
			wantErr: false,
		},
		"ignore $ when it is not escaped and no variable is followed by it": {
			args:    args{`FOO="bar $"`, false},
			want:    envVars{"FOO": "bar $"},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			for key, value := range tt.setEnvs {
				t.Setenv(key, value)
			}
			got, err := parseString(tt.args.contents, tt.args.overload)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir("./testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatal(err)
		}
	}(pwd)

	type args struct {
		options []LoadOption
	}
	tests := map[string]struct {
		args    args
		setEnvs envVars
		want    envVars
		wantErr bool
	}{
		"defaults to loading .env": {
			args:    args{},
			want:    envVars{"DOTENV": "true"},
			wantErr: false,
		},
		"loads nothing with empty file list": {
			args:    args{options: []LoadOption{Files()}},
			want:    envVars{},
			wantErr: false,
		},
		"loads nothing when the file does not exist": {
			args:    args{options: []LoadOption{Files(".env.does_not_exist")}},
			want:    envVars{},
			wantErr: false,
		},
		"load variables from file": {
			args:    args{options: []LoadOption{Files(".env")}},
			want:    envVars{"DOTENV": "true"},
			wantErr: false,
		},
		"load variables from multiple files": {
			args: args{options: []LoadOption{Files(".env", "plain.env")}},
			want: envVars{
				"PLAIN":    "true",
				"OPTION_A": "1",
				"OPTION_B": "2",
				"OPTION_C": "3",
				"OPTION_D": "4",
				"OPTION_E": "5",
				"DOTENV":   "true",
			},
			wantErr: false,
		},
		"load variables from file does not overwrite ENV": {
			args:    args{options: []LoadOption{Files(".env")}},
			setEnvs: envVars{"DOTENV": "false"},
			want:    envVars{"DOTENV": "false"},
			wantErr: false,
		},
		"load variables from multiple files does not overwrite ENV": {
			args:    args{options: []LoadOption{Files(".env", "plain.env")}},
			setEnvs: envVars{"OPTION_A": "predefined"},
			want: envVars{
				"PLAIN":    "true",
				"OPTION_A": "predefined",
				"OPTION_B": "2",
				"OPTION_C": "3",
				"OPTION_D": "4",
				"OPTION_E": "5",
				"DOTENV":   "true",
			},
			wantErr: false,
		},
		"returns an error when required files do not exist": {
			args:    args{options: []LoadOption{Files(".env", ".env.does_not_exist"), AllFilesRequired()}},
			want:    nil,
			wantErr: true,
		},
		"overload variables from file": {
			args:    args{options: []LoadOption{Files(".env"), Overload()}},
			setEnvs: envVars{"DOTENV": "false"},
			want:    envVars{"DOTENV": "true"},
			wantErr: false,
		},
		"overload variables from multiple files": {
			args:    args{options: []LoadOption{Files(".env", "plain.env"), Overload()}},
			setEnvs: envVars{"PLAIN": "false", "DOTENV": "false"},
			want: envVars{
				"PLAIN":    "true",
				"OPTION_A": "1",
				"OPTION_B": "2",
				"OPTION_C": "3",
				"OPTION_D": "4",
				"OPTION_E": "5",
				"DOTENV":   "true",
			},
			wantErr: false,
		},
		"present required keys pass": {
			args:    args{options: []LoadOption{Files(".env"), RequiredKeys("DOTENV")}},
			setEnvs: envVars{"DOTENV": "false"},
			want:    envVars{"DOTENV": "false"},
			wantErr: false,
		},
		"missing required keys returns an error": {
			args:    args{options: []LoadOption{Files(".env"), RequiredKeys("TEST")}},
			want:    nil,
			wantErr: true,
		},
		"load variables from files in multiple paths": {
			args:    args{options: []LoadOption{Paths(".", "nested")}},
			want:    envVars{"DOTENV": "true", "NESTED": "true"},
			wantErr: false,
		},
		"load variables for an environment 1": {
			args: args{options: []LoadOption{EnvironmentFiles("development")}},
			want: envVars{
				"DOTENV":                 "development-local",
				"DOTENVDEVELOPMENT":      "true",
				"DOTENVDEVELOPMENTLOCAL": "true",
				"DOTENVLOCAL":            "true",
			},
			wantErr: false,
		},
		"load variables for an environment 2": {
			args: args{options: []LoadOption{EnvironmentFiles("test")}},
			want: envVars{
				"DOTENV":     "test",
				"DOTENVTEST": "true",
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Clearenv()
			for key, value := range tt.setEnvs {
				t.Setenv(key, value)
			}
			err := Load(tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// skip testing the ENV; any issues we see after here will not be problems
			if tt.wantErr {
				return
			}
			envs := systemEnvs()
			if !reflect.DeepEqual(envs, tt.want) {
				t.Errorf("ENV = %v, want %v", envs, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir("./testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func(dir string) {
		err := os.Chdir(dir)
		if err != nil {
			t.Fatal(err)
		}
	}(pwd)

	type args struct {
		options []ParseOption
	}
	tests := map[string]struct {
		args    args
		setEnvs envVars
		want    map[string]string
		wantErr bool
	}{
		"defaults to loading .env": {
			args:    args{},
			want:    envVars{"DOTENV": "true"},
			wantErr: false,
		},
		"loads nothing with empty file list": {
			args:    args{options: []ParseOption{Files()}},
			want:    envVars{},
			wantErr: false,
		},
		"loads nothing when the file does not exist": {
			args:    args{options: []ParseOption{Files(".env.does_not_exist")}},
			want:    envVars{},
			wantErr: false,
		},
		"load variables from file": {
			args:    args{options: []ParseOption{Files(".env")}},
			want:    envVars{"DOTENV": "true"},
			wantErr: false,
		},
		"load variables from multiple files": {
			args: args{options: []ParseOption{Files(".env", "plain.env")}},
			want: envVars{
				"PLAIN":    "true",
				"OPTION_A": "1",
				"OPTION_B": "2",
				"OPTION_C": "3",
				"OPTION_D": "4",
				"OPTION_E": "5",
				"DOTENV":   "true",
			},
			wantErr: false,
		},
		"returns an error when required files do not exist": {
			args:    args{options: []ParseOption{Files(".env", ".env.does_not_exist"), AllFilesRequired()}},
			want:    nil,
			wantErr: true,
		},
		"load variables from files in multiple paths": {
			args:    args{options: []ParseOption{Paths(".", "nested")}},
			want:    envVars{"DOTENV": "true", "NESTED": "true"},
			wantErr: false,
		},
		"load variables for an environment 1": {
			args: args{options: []ParseOption{EnvironmentFiles("development")}},
			want: envVars{
				"DOTENV":                 "development-local",
				"DOTENVDEVELOPMENT":      "true",
				"DOTENVDEVELOPMENTLOCAL": "true",
				"DOTENVLOCAL":            "true",
			},
			wantErr: false,
		},
		"load variables for an environment 2": {
			args: args{options: []ParseOption{EnvironmentFiles("test")}},
			want: envVars{
				"DOTENV":     "test",
				"DOTENVTEST": "true",
			},
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			os.Clearenv()
			for key, value := range tt.setEnvs {
				t.Setenv(key, value)
			}
			got, err := Parse(tt.args.options...)
			// reset ENV

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
			}
			envs := systemEnvs()
			if len(envs) != 0 {
				t.Errorf("ENV got = %v, want %v", envs, envVars{})
			}
		})
	}
}
