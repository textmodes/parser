package teletext

type attr struct {
	char         uint8
	fg, bg       uint8
	charset      uint8
	doubleHeight bool
	doubleWidth  bool
	concealed    bool
	inverted     bool
	flashing     uint8
	diacritical  uint8
	underline    bool
	boxwin       bool
	setX26       bool
	setG0G2      uint8
}
