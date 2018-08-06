package server

import "testing"

func TestClientStateAdd(t *testing.T) {
	state := NewClientState()

	cases := []struct {
		name  string
		token string
		ok    bool
	}{
		{
			name:  "Alice",
			token: "Example",
			ok:    true,
		},
		{
			name:  "Alice",
			token: "Example",
			ok:    false,
		},
		{
			name:  "Alice",
			token: "Example2",
			ok:    false,
		},
		{
			name:  "Bob",
			token: "Example3",
			ok:    true,
		},
	}

	for _, tc := range cases {
		ok := state.Add(tc.name, tc.token)

		if tc.ok != ok {
			t.Errorf("Ok should be %t but got %t (%+v)", tc.ok, ok, tc)
		}
	}

}

func TestClientStateGetNameByToken(t *testing.T) {
	name := "Alice"
	token := "example"

	state := NewClientState()
	state.Add(name, token)

	cases := []struct {
		name  string
		ok    bool
		token string
	}{
		{
			token: "unknown",
			name:  "",
			ok:    false,
		},
		{
			token: token,
			name:  name,
			ok:    true,
		},
	}

	for _, tc := range cases {
		n, ok := state.GetNameByToken(tc.token)

		if tc.ok != ok {
			t.Errorf("Ok should be %t but got %t (%+v)", tc.ok, ok, tc)
		}

		if tc.name != n {
			t.Errorf("Name should be %s but got %s (%+v)", tc.name, n, tc)
		}
	}
}

func TestCLientStateRemove(t *testing.T) {
	fixtures := []struct {
		name  string
		token string
	}{
		{
			name:  "Alice",
			token: "example",
		},
		{
			name:  "Alice",
			token: "example2",
		},
		{
			name:  "Bob",
			token: "example3",
		},
	}

	cases := []struct {
		token string
		name  string
		ok    bool
	}{
		{
			token: "example",
			name:  "Alice",
			ok:    false,
		},
		{
			token: "example5",
			name:  "",
			ok:    false,
		},
		{
			token: "example2",
			name:  "Alice",
			ok:    true,
		},
		{
			token: "example3",
			name:  "Bob",
			ok:    true,
		},
		{
			token: "example6",
			name:  "",
			ok:    false,
		},
		{
			token: "example",
			name:  "",
			ok:    false,
		},
	}

	state := NewClientState()
	for _, item := range fixtures {
		state.Add(item.name, item.token)
	}

	for _, tc := range cases {
		name, ok := state.Remove(tc.token)

		if tc.ok != ok {
			t.Errorf("Ok should be %t but got %t (%+v)", tc.ok, ok, tc)
		}

		if tc.name != name {
			t.Errorf("Name should be %s but got %s (%+v)", tc.name, name, tc)
		}
	}
}

func TestClientStateAddStream(t *testing.T) {
	cases := []struct {
		token string
		len   int
	}{
		{
			token: "example",
			len:   1,
		},
		{
			token: "example2",
			len:   2,
		},
	}

	state := NewClientState()
	for _, tc := range cases {
		state.AddStream(tc.token)

		l := len(state.Streams)
		if tc.len != l {
			t.Errorf("Len should be %d but got %d (%+v)", tc.len, l, tc)
		}
	}
}

func TestClientStateCloseStream(t *testing.T) {
	state := NewClientState()

	fixtures := []string{"example", "example2"}

	for _, token := range fixtures {
		state.AddStream(token)
	}

	cases := []struct {
		token string
		len   int
	}{
		{
			token: "example",
			len:   1,
		},
		{
			token: "example2",
			len:   0,
		},
	}

	for _, tc := range cases {
		state.CloseStream(tc.token)

		l := len(state.Streams)
		if tc.len != l {
			t.Errorf("Len should be %d but got %d (%+v)", tc.len, l, tc)
		}
	}
}
