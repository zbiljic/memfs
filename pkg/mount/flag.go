// Package mount contains helper functions for dealing with mount(8)-style
// flags.
package mount

import "strings"

// ParseOptions an option string in the format accepted by mount(8) and
// generated for its external mount helpers.
//
// It is assumed that option name and values do not contain commas, and that
// the first equals sign in an option is the name/value separator. There is no
// support for escaping.
//
// For example, if the input is
//
//     user,foo=bar=baz,qux
//
// then the following will be inserted into the map.
//
//     "user": "",
//     "foo": "bar=baz",
//     "qux": "",
//
func ParseOptions(m map[string]string, s string) {
	// NOTE(jacobsa): The man pages don't define how escaping works, and as far
	// as I can tell there is no way to properly escape or quote a comma in the
	// options list for an fstab entry. So put our fingers in our ears and hope
	// that nobody needs a comma.
	for _, p := range strings.Split(s, ",") {
		var name string
		var value string

		// Split on the first equals sign.
		if equalsIndex := strings.IndexByte(p, '='); equalsIndex != -1 {
			name = p[:equalsIndex]
			value = p[equalsIndex+1:]
		} else {
			name = p
		}

		m[name] = value
	}

	return
}
