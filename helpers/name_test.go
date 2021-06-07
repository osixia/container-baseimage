package helpers

import (
	"testing"
)

func TestValidateName(t *testing.T) {

	for _, name := range []string{"nginx", "php-fpm", "my_service.2", "..hidden", "a..b"} {
		if err := ValidateName(name); err != nil {
			t.Errorf("%q: unexpected error: %v", name, err)
		}
	}

	for _, name := range []string{"", ".", "..", "../etc", "a/b", "/abs", `a\b`, "a\x00b"} {
		if err := ValidateName(name); err == nil {
			t.Errorf("%q: expected an error", name)
		}
	}

	if err := ValidateNames("nginx", "../etc"); err == nil {
		t.Error("ValidateNames must reject any invalid name")
	}
}

func TestValidateUnixID(t *testing.T) {

	for _, id := range []string{"0", "1000", "65534"} {
		if err := ValidateUnixID(id); err != nil {
			t.Errorf("%q: unexpected error: %v", id, err)
		}
	}

	for _, id := range []string{"", "-1", "1000 -o", "1000;id", "1e3", " 1000"} {
		if err := ValidateUnixID(id); err == nil {
			t.Errorf("%q: expected an error", id)
		}
	}
}

func TestValidateUnixName(t *testing.T) {

	for _, name := range []string{"nginx", "_apt", "www-data", "user1"} {
		if err := ValidateUnixName(name); err != nil {
			t.Errorf("%q: unexpected error: %v", name, err)
		}
	}

	for _, name := range []string{"", "-o", "1user", "a b", "a;id", "a$(id)", "a\nb", "a/b"} {
		if err := ValidateUnixName(name); err == nil {
			t.Errorf("%q: expected an error", name)
		}
	}
}
