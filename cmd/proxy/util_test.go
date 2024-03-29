package main

import (
	"fmt"
	"strings"
	"testing"
	"testing/quick"
)

func TestStringSlice(t *testing.T) {
	fn := func(a, b, c string) bool {
		var ss stringSlice
		ss.Set(a)
		ss.Set(b)
		ss.Set(c)

		if expected, actual := fmt.Sprintf("%s %s %s", a, b, c), strings.Join(ss, " "); expected != actual {
			t.Errorf("expected: %q, actual: %q", expected, actual)
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Error(err)
	}
}

func TestParseAddr(t *testing.T) {
	for _, testcase := range []struct {
		addr        string
		defaultPort int
		network     string
		address     string
	}{
		{"foo", 123, "tcp", "foo:123"},
		{"foo:80", 123, "tcp", "foo:80"},
		{"udp://foo", 123, "udp", "foo:123"},
		{"udp://foo:8080", 123, "udp", "foo:8080"},
		{"tcp+dnssrv://testing:7650", 7650, "tcp+dnssrv", "testing:7650"},
		{"[::]:", 123, "tcp", "0.0.0.0:123"},
		{"[::]:80", 123, "tcp", "0.0.0.0:80"},
	} {
		network, address, err := parseAddr(testcase.addr, testcase.defaultPort)
		if err != nil {
			t.Errorf("(%q, %d): %v", testcase.addr, testcase.defaultPort, err)
			continue
		}
		var (
			matchNetwork = network == testcase.network
			matchAddress = address == testcase.address
		)
		if !matchNetwork || !matchAddress {
			t.Errorf("(%q, %d): want [%s %s], have [%s %s]",
				testcase.addr, testcase.defaultPort,
				testcase.network, testcase.address,
				network, address,
			)
			continue
		}
	}
}
