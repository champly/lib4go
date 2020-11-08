package env

import (
	"os"
	"testing"
	"time"
)

const testVar = "TEST_VAR"

func reset() {
	_ = os.Unsetenv(testVar)

	mutex.Lock()
	allVar = make(map[string]Var)
	mutex.Unlock()
}

func TestLookup(t *testing.T) {
	// String
	reset()
	evs := RegistryStringVar(testVar, "123", "description")
	vs, ok := evs.Lookup()
	if vs != "123" {
		t.Errorf("Expected 123, got %s", vs)
	}
	if ok {
		t.Errorf("Expected not present")
	}

	vs = evs.Get()
	if vs != "123" {
		t.Errorf("expected 123, got %s", vs)
	}

	// Bool
	reset()
	evb := RegistryBoolVar(testVar, "t", "description")
	vb, ok := evb.Lookup()
	if !vb {
		t.Errorf("Expected true, got %t", vb)
	}
	if ok {
		t.Errorf("Expected not present")
	}
	vb = evb.Get()
	if !vb {
		t.Errorf("Expected true, got %t", vb)
	}
	// exception
	reset()
	evb = RegistryBoolVar(testVar, "abc", "description")
	vb = evb.Get()
	if vb {
		t.Errorf("Expected false, got %t", vb)
	}

	// Float
	reset()
	evf := RegistryFloatVar(testVar, "1.2", "description")
	vf, ok := evf.Lookup()
	if vf != 1.2 {
		t.Errorf("Expected 1.2, got %f", vf)
	}
	if ok {
		t.Errorf("Expected not present")
	}
	vf = evf.Get()
	if vf != 1.2 {
		t.Errorf("Expected 1.2, got %f", vf)
	}
	// exception
	reset()
	evf = RegistryFloatVar(testVar, "abc", "descrition")
	vf = evf.Get()
	if vf != 0 {
		t.Errorf("Expected 0, got %f", vf)
	}

	// Int
	reset()
	evi := RegistryIntVar(testVar, "111", "description")
	vi, ok := evi.Lookup()
	if vi != 111 {
		t.Errorf("Expected 111, got %d", vi)
	}
	if ok {
		t.Errorf("Expected not present")
	}
	vi = evi.Get()
	if vi != 111 {
		t.Errorf("Expected 111, got %d", vi)
	}
	// exception
	reset()
	evi = RegistryIntVar(testVar, "abc", "decription")
	vi = evi.Get()
	if vi != 0 {
		t.Errorf("Expected 0, got %d", vi)
	}

	// Duration
	reset()
	evd := RegistryDurationVar(testVar, "1s", "description")
	vd, ok := evd.Lookup()
	if vd != time.Second {
		t.Errorf("Expected 1s, got %v", vd)
	}
	if ok {
		t.Errorf("Expected not present")
	}
	vd = evd.Get()
	if vd != time.Second {
		t.Errorf("Expected 1s, got %v", vd)
	}
	// exception
	reset()
	evd = RegistryDurationVar(testVar, "abc", "description")
	vd = evd.Get()
	if vd != 0 {
		t.Errorf("Expected 0， got %v", vd)
	}
}

func TestRepeatRegistry(t *testing.T) {
	reset()
	RegistryBoolVar(testVar, "true", "")
	RegistryFloatVar(testVar, "1.2", "")
}
