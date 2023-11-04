package args

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
)

type FlagSet struct {
	formal map[string]*Flag
	actual map[string]struct{}

	envPrefix *string
	parsers   []Parse
}

type Flag struct {
	Long     string
	Short    string
	Usage    string
	Value    Value
	DefValue string
}

func New() *FlagSet {
	return &FlagSet{
		formal: make(map[string]*Flag),
		actual: make(map[string]struct{}),
	}
}

func (f *FlagSet) String(name string, value string, usage string) *string {
	v := newStringValue(value)
	f.Add(v, name, usage)
	return (*string)(v)
}

func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	v := newBoolValue(value)
	f.Add(v, name, usage)
	return (*bool)(v)
}

func (f *FlagSet) Int(name string, value int, usage string) *int {
	v := newIntValue(value)
	f.Add(v, name, usage)
	return (*int)(v)
}

func (f *FlagSet) Int64(name string, value int64, usage string) *int64 {
	v := newInt64Value(value)
	f.Add(v, name, usage)
	return (*int64)(v)
}

func (f *FlagSet) Uint(name string, value uint, usage string) *uint {
	v := newUintValue(value)
	f.Add(v, name, usage)
	return (*uint)(v)
}

func (f *FlagSet) Uint64(name string, value uint64, usage string) *uint64 {
	v := newUint64Value(value)
	f.Add(v, name, usage)
	return (*uint64)(v)
}

func (f *FlagSet) Duration(name string, value time.Duration, usage string) *time.Duration {
	v := newDuration(value)
	f.Add(v, name, usage)
	return (*time.Duration)(v)
}

func (f *FlagSet) Add(value Value, name string, usage string) {
	long, short, err := parseName(name)
	if err != nil {
		panic(err)
	}

	flag := Flag{
		Long:     long,
		Short:    short,
		Usage:    usage,
		Value:    value,
		DefValue: value.String(),
	}

	if flag.Long != "" {
		if _, exists := f.formal[flag.Long]; exists {
			panic("long flag redefined: " + flag.Long)
		}
		f.formal[flag.Long] = &flag
	}

	if flag.Short != "" {
		if _, exists := f.formal[flag.Short]; exists {
			panic("short flag redefined: " + flag.Short)
		}
		f.formal[flag.Short] = &flag
	}
}

func (f *FlagSet) addActual(flag *Flag) {
	if flag.Long != "" {
		f.actual[flag.Long] = struct{}{}
	}
	if flag.Short != "" {
		f.actual[flag.Short] = struct{}{}
	}
}

// parses a flag name, which can be in the format "host", "h", "host,h" or "h,host"
func parseName(name string) (long string, short string, err error) {
	parts := strings.Split(name, ",")
	switch len(parts) {
	case 1:
		long = parts[0]
	case 2:
		long, short = parts[0], parts[1]
	default:
		return "", "", fmt.Errorf("invalid flag name format: %s", name)
	}

	if len(long) == 1 {
		long, short = short, long
	}

	if long != "" {
		if len(long) == 1 {
			return "", "", fmt.Errorf("long flag %q should be a word", long)
		}
		if strings.HasPrefix(long, "-") {
			return "", "", fmt.Errorf("long flag %q begins with -", long)
		}
		if strings.Contains(long, "=") {
			return "", "", fmt.Errorf("long flag %q contains =", long)
		}
		if strings.Contains(long, " ") {
			return "", "", fmt.Errorf("long flag %q contains space", long)
		}
	}

	if short != "" {
		if len(short) != 1 || !unicode.IsLetter(rune(short[0])) {
			return "", "", fmt.Errorf("short flag %q should be a letter", short)
		}
	}

	return long, short, nil
}

type Parse = func(set func(name, value string) error) error

func (f *FlagSet) Parse(arguments []string, options ...option) {
	for i := range options {
		options[i](f)
	}

	if err := f.parseCommandLine(arguments); err != nil {
		panic(err)
	}
	if err := f.parseEnvironment(); err != nil {
		panic(err)
	}

	for _, parser := range f.parsers {
		err := parser(func(name string, value string) error {
			if _, exists := f.actual[name]; exists {
				return nil
			}

			flag, exists := f.formal[name]
			if !exists {
				return nil
			}
			if err := flag.Value.Set(value); err != nil {
				return err
			}

			f.addActual(flag)

			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}

func (f *FlagSet) parseCommandLine(arguments []string) error {
	var cursor int
	for cursor < len(arguments) {
		arg := arguments[cursor]
		if len(arg) < 2 || arg[0] != '-' {
			return nil
		}

		// "--" terminates the flags
		if arg == "--" {
			return nil
		}

		isLongFlag := arg[1] == '-'
		if isLongFlag {
			arg = arg[2:]
		} else {
			arg = arg[1:]
		}

		if len(arg) == 0 || arg[0] == '-' || arg[0] == '=' {
			return fmt.Errorf("bad flag syntax: %s", arg)
		}

		// it's a flag. does it have an argument?
		var value string
		for i := 1; i < len(arg); i++ {
			if arg[i] == '=' {
				arg, value = arg[:i], arg[i+1:]
				break
			}
		}

		// short flag supports multiple boolean flags format like "-abc" means "-a -b -c"
		if !isLongFlag && len(arg) != 1 {
			parts := strings.Split(arg, "")
			for _, part := range parts {
				flag, exists := f.formal[part]
				if !exists {
					return fmt.Errorf("flag provide but not defined: %s", part)
				}
				if v, ok := flag.Value.(BoolValue); !ok || !v.IsBoolValue() {
					return fmt.Errorf("flag not boolean flag: %s", part)
				}
				if err := flag.Value.Set("true"); err != nil {
					return fmt.Errorf("invalid boolean flag: %s: %v", part, err)
				}

				f.addActual(flag)
			}
		} else {
			flag, exists := f.formal[arg]
			if !exists {
				if arg == "help" || arg == "h" {
					return nil
				}
				return fmt.Errorf("flag provided but not defined: %s", arg)
			}

			if v, ok := flag.Value.(BoolValue); ok && v.IsBoolValue() {
				if value == "" && cursor+1 < len(arguments) && arguments[cursor+1][0] != '-' {
					value = arguments[cursor+1]
					cursor++
				}
				if value != "" {
					if err := flag.Value.Set(value); err != nil {
						return fmt.Errorf("invalid boolean value %q for %s: %v", value, arg, err)
					}
				} else {
					if err := flag.Value.Set("true"); err != nil {
						return fmt.Errorf("invalid boolean flag %s: %v", arg, err)
					}
				}
			} else {
				if value == "" && cursor+1 < len(arguments) {
					value = arguments[cursor+1]
					cursor++
				}
				if value == "" {
					return fmt.Errorf("flag needs an arguments: %s", arg)
				}
				if err := flag.Value.Set(value); err != nil {
					return fmt.Errorf("invalid value %q for flag: %s: %v", value, arg, err)
				}
			}

			f.addActual(flag)
		}

		cursor++
	}

	return nil
}

func (f *FlagSet) parseEnvironment() error {
	if f.envPrefix == nil {
		return nil
	}

	replacer := strings.NewReplacer(
		"-", "_",
		".", "_",
		"/", "_",
	)

	prefix := strings.ToUpper(*f.envPrefix)
	for name, flag := range f.formal {
		if _, exists := f.actual[name]; exists {
			continue
		}

		// ignore short flag
		if flag.Long == "" {
			continue
		}

		key := replacer.Replace(strings.ToUpper(flag.Long))
		if prefix != "" {
			key = prefix + key
		}

		value := os.Getenv(key)
		if value == "" {
			continue
		}

		if err := flag.Value.Set(value); err != nil {
			return err
		}

		f.addActual(flag)
	}

	return nil
}
