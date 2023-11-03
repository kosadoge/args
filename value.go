package args

import (
	"strconv"
	"time"
)

type Value interface {
	String() string
	Set(string) error
}

type BoolValue interface {
	Value
	IsBoolValue() bool
}

type stringValue string

func newStringValue(val string) *stringValue { return (*stringValue)(&val) }
func (s *stringValue) String() string        { return string(*s) }
func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

type boolValue bool

func newBoolValue(val bool) *boolValue { return (*boolValue)(&val) }
func (b *boolValue) IsBoolValue() bool { return true }
func (b *boolValue) String() string    { return strconv.FormatBool(bool(*b)) }
func (b *boolValue) Set(val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	*b = boolValue(v)
	return nil
}

type intValue int

func newIntValue(val int) *intValue { return (*intValue)(&val) }
func (i *intValue) String() string  { return strconv.Itoa(int(*i)) }
func (i *intValue) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return nil
}

type durationValue time.Duration

func newDuration(val time.Duration) *durationValue { return (*durationValue)(&val) }
func (d *durationValue) String() string            { return time.Duration(*d).String() }
func (d *durationValue) Set(val string) error {
	v, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	*d = durationValue(v)
	return nil
}
