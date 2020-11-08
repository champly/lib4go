package env

import (
	"os"
	"strconv"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

// VarType ...
type VarType byte

const (
	STRING VarType = iota
	BOOL
	INT
	FLOAT
	DURATION
)

// Var ...
type Var struct {
	Name         string
	DefaultValue string
	Description  string
	Hidden       bool
	Deprecated   bool
	Type         VarType
}

var (
	allVar = make(map[string]Var)
	mutex  sync.Mutex
)

func registryVar(v Var) {
	mutex.Lock()
	defer mutex.Unlock()

	if old, ok := allVar[v.Name]; ok {
		if old.DefaultValue != v.DefaultValue || old.Description != v.Description || old.Hidden != v.Hidden || old.Deprecated != v.Deprecated || old.Type != v.Type {
			klog.Warningf("The enviroment variable %s was registered multiple times using different metadata: %v %v", v.Name, old, v)
			allVar[v.Name] = v
		}
	} else {
		allVar[v.Name] = v
	}
}

func getVar(name string) Var {
	mutex.Lock()
	defer mutex.Unlock()

	return allVar[name]
}

// StringVar ...
type StringVar struct {
	Var
}

// RegistryStringVar ...
func RegistryStringVar(name string, defaultValue string, description string) StringVar {
	v := Var{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
		Type:         STRING,
	}
	registryVar(v)
	return StringVar{getVar(name)}
}

func (v StringVar) Get() string {
	result, _ := v.Lookup()
	return result
}

func (v StringVar) Lookup() (string, bool) {
	result, ok := os.LookupEnv(v.Name)
	if !ok {
		result = v.DefaultValue
	}
	return result, ok
}

// BoolVar ...
type BoolVar struct {
	Var
}

// RegistryBoolVar ...
func RegistryBoolVar(name string, defaultValue string, description string) BoolVar {
	v := Var{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
		Type:         BOOL,
	}
	registryVar(v)
	return BoolVar{getVar(name)}
}

func (v BoolVar) Get() bool {
	result, _ := v.Lookup()
	return result
}

func (v BoolVar) Lookup() (bool, bool) {
	result, ok := os.LookupEnv(v.Name)
	if !ok {
		result = v.DefaultValue
	}
	b, err := strconv.ParseBool(result)
	if err != nil {
		klog.Warningf("Invalid environment variable value `%s`, expecting true/false, defaulting to %v", result, v.DefaultValue)
		b, _ = strconv.ParseBool(v.DefaultValue)
	}
	return b, ok
}

// IntVar ...
type IntVar struct {
	Var
}

// RegistryIntVar ...
func RegistryIntVar(name string, defaultValue string, description string) IntVar {
	v := Var{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
		Type:         INT,
	}
	registryVar(v)
	return IntVar{getVar(name)}
}

func (v IntVar) Get() int {
	result, _ := v.Lookup()
	return result
}

func (v IntVar) Lookup() (int, bool) {
	result, ok := os.LookupEnv(v.Name)
	if !ok {
		result = v.DefaultValue
	}
	b, err := strconv.Atoi(result)
	if err != nil {
		klog.Warningf("Invalid environment variable value `%s`, expecting an integer, defaulting to %v", result, v.DefaultValue)
		b, _ = strconv.Atoi(v.DefaultValue)
	}
	return b, ok
}

// FloatVar ...
type FloatVar struct {
	Var
}

// RegistryFloatVar ...
func RegistryFloatVar(name string, defaultValue string, description string) FloatVar {
	v := Var{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
		Type:         FLOAT,
	}
	registryVar(v)
	return FloatVar{getVar(name)}
}

func (v FloatVar) Get() float64 {
	result, _ := v.Lookup()
	return result
}

func (v FloatVar) Lookup() (float64, bool) {
	result, ok := os.LookupEnv(v.Name)
	if !ok {
		result = v.DefaultValue
	}
	b, err := strconv.ParseFloat(result, 64)
	if err != nil {
		klog.Warningf("Invalid environment variable value `%s`, expecting floating-point value, defaulting to %v", result, v.DefaultValue)
		b, _ = strconv.ParseFloat(v.DefaultValue, 64)
	}
	return b, ok
}

// DurationVar ...
type DurationVar struct {
	Var
}

// RegistryDurationVar ...
func RegistryDurationVar(name string, defaultValue string, description string) DurationVar {
	v := Var{
		Name:         name,
		DefaultValue: defaultValue,
		Description:  description,
		Type:         DURATION,
	}
	registryVar(v)
	return DurationVar{getVar(name)}
}

func (v DurationVar) Get() time.Duration {
	result, _ := v.Lookup()
	return result
}

func (v DurationVar) Lookup() (time.Duration, bool) {
	result, ok := os.LookupEnv(v.Name)
	if !ok {
		result = v.DefaultValue
	}

	d, err := time.ParseDuration(result)
	if err != nil {
		klog.Warningf("Invalid environment variable value `%s`, expecting a duration, defaulting to %v", result, v.DefaultValue)
		d, _ = time.ParseDuration(v.DefaultValue)
	}
	return d, ok
}
