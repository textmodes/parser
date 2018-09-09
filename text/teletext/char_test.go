package teletext

import "testing"

func TestCharMap(t *testing.T) {
	var tests = map[string]string{
		"Simple stuff":   "Simple stuff",
		"The fish is #2": "The fish is £2", // '#':  0x00A3, // 2/3 # is mapped to pound sign
		"Right ]here[":   "Right →here←",   // '[':  0x2190, // 5/B Left arrow.
		"Up there ^":     "Up there ↑",     // '^':  0x2191, // 5/E Up arrow.
		"\\ of the pie":  "½ of the pie",   // '\\': 0x00bd, // 5/C Half
		"_twitter":       "#twitter",       // '_':  0x0023, // 5/F Underscore is hash sign
		"center ` dash":  "center — dash",  // '`':  0x2014, // 6/0 Centre dash. The full width dash e731
		"{ of the pie":   "¼ of the pie",   // '{':  0x00bc, // 7/B Quarter
		"pipe | [there":  "pipe ‖ ←there",  // '|':  0x2016, // 7/C Double pipe
		"} of the pie":   "¾ of the pie",   // '}':  0x00be, // 7/D Three quarters
		"1~2 = \\":       "1÷2 = ½",        // '~':  0x00f7, // 7/E Divide
	}
	for test, want := range tests {
		t.Run(test, func(t *testing.T) {
			if got := regions[0][English].String(test); got != want {
				t.Fatalf("expected %q to return %q, got %q", test, want, got)
			}
		})
	}
}
