// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package ptr

import (
	"testing"
)

func TestTo(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		v := 42
		p := To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Errorf("To(%d) = %d, want %d", v, *p, v)
		}
	})

	t.Run("string", func(t *testing.T) {
		v := "hello"
		p := To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Errorf("To(%q) = %q, want %q", v, *p, v)
		}
	})

	t.Run("bool", func(t *testing.T) {
		v := true
		p := To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Errorf("To(%v) = %v, want %v", v, *p, v)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type myStruct struct {
			A int
			B string
		}
		v := myStruct{A: 1, B: "test"}
		p := To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Errorf("To(%v) = %v, want %v", v, *p, v)
		}
	})
}

func TestDeref(t *testing.T) {
	t.Run("int non-nil", func(t *testing.T) {
		v := 42
		result := Deref(&v, 0)
		if result != v {
			t.Errorf("Deref(&%d, 0) = %d, want %d", v, result, v)
		}
	})

	t.Run("int nil", func(t *testing.T) {
		var p *int
		result := Deref(p, 99)
		if result != 99 {
			t.Errorf("Deref(nil, 99) = %d, want 99", result)
		}
	})

	t.Run("string non-nil", func(t *testing.T) {
		v := "hello"
		result := Deref(&v, "default")
		if result != v {
			t.Errorf("Deref(&%q, \"default\") = %q, want %q", v, result, v)
		}
	})

	t.Run("string nil", func(t *testing.T) {
		var p *string
		result := Deref(p, "default")
		if result != "default" {
			t.Errorf("Deref(nil, \"default\") = %q, want \"default\"", result)
		}
	})

	t.Run("bool non-nil true", func(t *testing.T) {
		v := true
		result := Deref(&v, false)
		if result != true {
			t.Errorf("Deref(&true, false) = %v, want true", result)
		}
	})

	t.Run("bool non-nil false", func(t *testing.T) {
		v := false
		result := Deref(&v, true)
		if result != false {
			t.Errorf("Deref(&false, true) = %v, want false", result)
		}
	})

	t.Run("bool nil", func(t *testing.T) {
		var p *bool
		result := Deref(p, true)
		if result != true {
			t.Errorf("Deref(nil, true) = %v, want true", result)
		}
	})

	t.Run("int32", func(t *testing.T) {
		v := int32(42)
		result := Deref(&v, int32(0))
		if result != v {
			t.Errorf("Deref(&%d, 0) = %d, want %d", v, result, v)
		}
	})

	t.Run("int64", func(t *testing.T) {
		v := int64(9999999999)
		result := Deref(&v, int64(0))
		if result != v {
			t.Errorf("Deref(&%d, 0) = %d, want %d", v, result, v)
		}
	})
}

func TestEqual(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		var a, b *int
		if !Equal(a, b) {
			t.Error("Equal(nil, nil) = false, want true")
		}
	})

	t.Run("first nil", func(t *testing.T) {
		var a *int
		b := To(42)
		if Equal(a, b) {
			t.Error("Equal(nil, &42) = true, want false")
		}
	})

	t.Run("second nil", func(t *testing.T) {
		a := To(42)
		var b *int
		if Equal(a, b) {
			t.Error("Equal(&42, nil) = true, want false")
		}
	})

	t.Run("both non-nil equal", func(t *testing.T) {
		a := To(42)
		b := To(42)
		if !Equal(a, b) {
			t.Error("Equal(&42, &42) = false, want true")
		}
	})

	t.Run("both non-nil different", func(t *testing.T) {
		a := To(42)
		b := To(99)
		if Equal(a, b) {
			t.Error("Equal(&42, &99) = true, want false")
		}
	})

	t.Run("string equal", func(t *testing.T) {
		a := To("hello")
		b := To("hello")
		if !Equal(a, b) {
			t.Error("Equal(&\"hello\", &\"hello\") = false, want true")
		}
	})

	t.Run("string different", func(t *testing.T) {
		a := To("hello")
		b := To("world")
		if Equal(a, b) {
			t.Error("Equal(&\"hello\", &\"world\") = true, want false")
		}
	})
}

func TestAllPtrFieldsNil(t *testing.T) {
	type TestStruct struct {
		A *int
		B *string
		C int // non-pointer field
	}

	t.Run("all pointer fields nil", func(t *testing.T) {
		s := TestStruct{C: 42}
		if !AllPtrFieldsNil(s) {
			t.Error("AllPtrFieldsNil returned false for struct with all nil pointer fields")
		}
	})

	t.Run("one pointer field non-nil", func(t *testing.T) {
		v := 42
		s := TestStruct{A: &v, C: 42}
		if AllPtrFieldsNil(s) {
			t.Error("AllPtrFieldsNil returned true for struct with non-nil pointer field")
		}
	})

	t.Run("struct pointer nil", func(t *testing.T) {
		var s *TestStruct
		if !AllPtrFieldsNil(s) {
			t.Error("AllPtrFieldsNil returned false for nil struct pointer")
		}
	})

	t.Run("struct pointer with nil fields", func(t *testing.T) {
		s := &TestStruct{C: 42}
		if !AllPtrFieldsNil(s) {
			t.Error("AllPtrFieldsNil returned false for struct pointer with all nil pointer fields")
		}
	})

	t.Run("struct pointer with non-nil field", func(t *testing.T) {
		v := 42
		s := &TestStruct{A: &v}
		if AllPtrFieldsNil(s) {
			t.Error("AllPtrFieldsNil returned true for struct pointer with non-nil pointer field")
		}
	})
}

// Fuzz tests

func FuzzDerefInt(f *testing.F) {
	f.Add(0, 99)
	f.Add(42, 0)
	f.Add(-1, 100)
	f.Add(1000000, -1000000)

	f.Fuzz(func(t *testing.T, val, def int) {
		// Non-nil case
		result := Deref(&val, def)
		if result != val {
			t.Errorf("Deref(&%d, %d) = %d, want %d", val, def, result, val)
		}

		// Nil case
		var nilPtr *int
		result = Deref(nilPtr, def)
		if result != def {
			t.Errorf("Deref(nil, %d) = %d, want %d", def, result, def)
		}
	})
}

func FuzzDerefString(f *testing.F) {
	f.Add("hello", "default")
	f.Add("", "empty")
	f.Add("test", "")
	f.Add("日本語", "unicode")

	f.Fuzz(func(t *testing.T, val, def string) {
		// Non-nil case
		result := Deref(&val, def)
		if result != val {
			t.Errorf("Deref(&%q, %q) = %q, want %q", val, def, result, val)
		}

		// Nil case
		var nilPtr *string
		result = Deref(nilPtr, def)
		if result != def {
			t.Errorf("Deref(nil, %q) = %q, want %q", def, result, def)
		}
	})
}

func FuzzEqual(f *testing.F) {
	f.Add(0, 0)
	f.Add(42, 42)
	f.Add(42, 99)
	f.Add(-1, 1)

	f.Fuzz(func(t *testing.T, a, b int) {
		ptrA := To(a)
		ptrB := To(b)

		result := Equal(ptrA, ptrB)
		expected := a == b

		if result != expected {
			t.Errorf("Equal(&%d, &%d) = %v, want %v", a, b, result, expected)
		}
	})
}
