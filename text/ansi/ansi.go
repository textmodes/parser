package ansi

func isdigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// ASCII escape characters
const (
	NUL = iota
	SOH
	STX
	ETX
	EOT
	ENQ
	ACK
	BEL
	BS
	TAB
	LF
	VT
	FF
	CR
	SO
	SI
	DLE
	DC1
	DC2
	DC3
	DC4
	NAK
	SYN
	ETB
	CAN
	EM
	SUB
	ESC
	FS
	GS
	RS
	US
	Space

	NL = LF
	NP = FF
)
