package teletext

type charMap map[rune]rune

// Rune translates one rune.
func (m charMap) Rune(r rune) rune {
	if o, ok := m[r]; ok {
		return o
	}
	return r
}

// String translates a string.
func (m charMap) String(s string) string {
	var (
		b = []byte(s)
		o = make([]rune, len(b))
	)
	for i, c := range b {
		o[i] = m.Rune(rune(c))
	}
	return string(o)
}

// Language code.
type Language int

func (lang Language) String() string {
	switch lang {
	case English:
		return "en"
	case French:
		return "fr"
	case Swedish:
		return "se"
	case Czech:
		return "cz/si"
	case German:
		return "de"
	case Spanish:
		return "es/pt"
	case Italian:
		return "it"
	default:
		return "unknown"
	}
}

// Languages.
const (
	English Language = iota
	French
	Swedish
	Czech
	German
	Spanish
	Italian

	// Aliases
	Slovac     = Czech
	Portuguese = Spanish
)

var regions = map[int]map[Language]charMap{
	0: map[Language]charMap{
		English: charMap{
			'#':  0x00A3, // 2/3 # is mapped to pound sign
			'$':  0x0024, // 2/4 Dollar sign (no change!)
			'@':  0x0040, // 4/0 No change
			'[':  0x2190, // 5/B Left arrow.
			'\\': 0x00bd, // 5/C Half
			']':  0x2192, // 5/D Right arrow.
			'^':  0x2191, // 5/E Up arrow.
			'_':  0x0023, // 5/F Underscore is hash sign
			'`':  0x2014, // 6/0 Centre dash. The full width dash e731
			'{':  0x00bc, // 7/B Quarter
			'|':  0x2016, // 7/C Double pipe
			'}':  0x00be, // 7/D Three quarters
			'~':  0x00f7, // 7/E Divide
		},
		French: charMap{
			'#':  0x00e9, // 2/3 e acute
			'$':  0x00ef, // 2/4 i umlaut
			'@':  0x00e0, // 4/0 a grave
			'[':  0x00eb, // 5/B e umlaut
			'\\': 0x00ea, // 5/C e circumflex
			']':  0x00f9, // 5/D u grave
			'^':  0x00ee, // 5/E i circumflex
			'_':  '#',    // 5/F #
			'`':  0x00e8, // 6/0 e grave
			'{':  0x00e2, // 7/B a circumflex
			'|':  0x00f4, // 7/C o circumflex
			'}':  0x00fb, // 7/D u circumflex
			'~':  0x00e7, // 7/E c cedilla
		},
	},
}
