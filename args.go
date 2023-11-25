package args

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"
)

type FlagSet struct {
	formal map[string]*Flag
	actual map[string]struct{}

	envPrefix *string
	providers []Provider
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
		long:     long,
		short:    short,
		usage:    usage,
		value:    value,
		defValue: value.String(),
	}

	if flag.long != "" {
		if _, exists := f.formal[flag.long]; exists {
			panic("long flag redefined: " + flag.long)
		}
		f.formal[flag.long] = &flag
	}

	if flag.short != "" {
		if _, exists := f.formal[flag.short]; exists {
			panic("short flag redefined: " + flag.short)
		}
		f.formal[flag.short] = &flag
	}
}

func (f *FlagSet) addActual(flag *Flag) {
	if flag.long != "" {
		f.actual[flag.long] = struct{}{}
	}
	if flag.short != "" {
		f.actual[flag.short] = struct{}{}
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

type Provider = func(set func(name, value string) error) error

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

	for _, provider := range f.providers {
		err := provider(func(name string, value string) error {
			if _, exists := f.actual[name]; exists {
				return nil
			}

			flag, exists := f.formal[name]
			if !exists {
				return nil
			}
			if err := flag.value.Set(value); err != nil {
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

		isShortFlag := arg[1] != '-'
		if isShortFlag {
			arg = arg[1:]
		} else {
			arg = arg[2:]
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
		if isShortFlag && len(arg) != 1 {
			parts := strings.Split(arg, "")
			for _, part := range parts {
				flag, exists := f.formal[part]
				if !exists {
					return fmt.Errorf("flag provide but not defined: %s", part)
				}
				if v, ok := flag.value.(BoolValue); !ok || !v.IsBoolValue() {
					return fmt.Errorf("flag not boolean flag: %s", part)
				}
				if err := flag.value.Set("true"); err != nil {
					return fmt.Errorf("invalid boolean flag: %s: %v", part, err)
				}

				f.addActual(flag)
			}
		} else {
			flag, exists := f.formal[arg]
			if !exists {
				if arg == "help" || arg == "h" {
					f.printUsage()
					return nil
				}
				return fmt.Errorf("flag provided but not defined: %s", arg)
			}

			if v, ok := flag.value.(BoolValue); ok && v.IsBoolValue() {
				if value == "" && cursor+1 < len(arguments) && arguments[cursor+1][0] != '-' {
					value = arguments[cursor+1]
					cursor++
				}
				if value != "" {
					if err := flag.value.Set(value); err != nil {
						return fmt.Errorf("invalid boolean value %q for %s: %v", value, arg, err)
					}
				} else {
					if err := flag.value.Set("true"); err != nil {
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
				if err := flag.value.Set(value); err != nil {
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

	prefix := *f.envPrefix
	for name, flag := range f.formal {
		if _, exists := f.actual[name]; exists {
			continue
		}

		// ignore short flag
		if flag.long == "" {
			continue
		}

		key := replacer.Replace(strings.ToUpper(flag.long))
		if prefix != "" {
			key = prefix + key
		}

		value := os.Getenv(key)
		if value == "" {
			continue
		}

		if err := flag.value.Set(value); err != nil {
			return err
		}

		f.addActual(flag)
	}

	return nil
}

func (f *FlagSet) printUsage() {
	type Line struct {
		flagName string
		usage    string
		defValue string
	}

	var maxLen int
	flags := sortFlags(f.formal)
	lines := make([]Line, 0, len(flags))
	for _, flag := range flags {
		var b strings.Builder
		hasShort, hasLong := flag.short != "", flag.long != ""
		switch {
		case hasShort && hasLong:
			fmt.Fprintf(&b, "  -%s, --%s", flag.short, flag.long)
		case hasShort && !hasLong:
			fmt.Fprintf(&b, "  -%s", flag.short)
		case !hasShort && hasLong:
			fmt.Fprintf(&b, "      --%s", flag.long)
		default:
			panic("unexpect flag name case")
		}

		if b.Len() > maxLen {
			maxLen = b.Len()
		}

		var defValue string
		isZero, err := isZeroValue(flag, flag.defValue)
		if err != nil {
			panic(err)
		}
		if !isZero {
			if _, ok := flag.value.(*stringValue); ok {
				defValue = fmt.Sprintf("%q", flag.defValue)
			} else {
				defValue = flag.defValue
			}
		}

		lines = append(lines, Line{
			flagName: b.String(),
			usage:    flag.usage,
			defValue: defValue,
		})
	}

	for _, l := range lines {
		if gap := maxLen - len(l.flagName); gap > 0 {
			l.flagName = l.flagName + strings.Repeat(" ", gap)
		}

		if l.defValue != "" {
			fmt.Printf("%s\t%s (default %s)\n", l.flagName, l.usage, l.defValue)
		} else {
			fmt.Printf("%s\t%s\n", l.flagName, l.usage)
		}
	}
}

func sortFlags(src map[string]*Flag) []*Flag {
	actual := make(map[string]struct{})
	flags := make([]*Flag, 0, len(src))
	for _, flag := range src {
		if _, exists := actual[flag.long]; exists {
			continue
		}
		if _, exists := actual[flag.short]; exists {
			continue
		}

		flags = append(flags, flag)
		if flag.long != "" {
			actual[flag.long] = struct{}{}
		}
		if flag.short != "" {
			actual[flag.short] = struct{}{}
		}
	}

	sort.Slice(flags, func(i, j int) bool {
		iname := flags[i].long
		if iname == "" {
			iname = flags[i].short
		}
		jname := flags[j].long
		if jname == "" {
			jname = flags[j].short
		}
		return iname < jname
	})

	return flags
}

func isZeroValue(flag *Flag, value string) (ok bool, err error) {
	typ := reflect.TypeOf(flag.value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}

	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Pointer {
				typ = typ.Elem()
			}

			var name string
			if flag.long != "" {
				name = flag.long
			} else {
				name = flag.short
			}
			err = fmt.Errorf("panic calling String method on zero %v for flag %s: %v", typ, name, e)
		}
	}()

	return value == z.Interface().(Value).String(), nil
}
