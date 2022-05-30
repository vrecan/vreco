package codingexercises

import (
	"testing"

	"golang.org/x/example/stringutil"
)

func TestReverseString(t *testing.T) {
	testString := "Abcdefg 汉语 The God ⚽😎 nwòrb"
	expected := stringutil.Reverse(testString)
	result := reverseString(testString)
	if result != expected {
		t.Errorf("Expected %s got %s", expected, result)
	}
}

func TestInplaceReverseString(t *testing.T) {
	testString := "Abcdefg 汉语 The God ⚽😎 nwòrb"
	expected := stringutil.Reverse(testString)
	result := inplaceReverseString(testString)
	if result != expected {
		t.Errorf("Expected %s got %s", expected, result)
	}
}
