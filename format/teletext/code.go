package teletext

// TeleText Level 1 character codes
const (
	CodeAlphaBlack         byte = iota // <-- Not teletext level 1
	CodeAlphaRed                       // Shift-F1
	CodeAlphaGreen                     // Shift-F2
	CodeAlphaYellow                    // Shift-F3
	CodeAlphaBlue                      // Shift-F4
	CodeAlphaMagenta                   // Shift-F5
	CodeAlphaCyan                      // Shift-F6
	CodeAlphaWhite                     // Shift-F7
	CodeFlash                          // Ctrl-H
	CodeSteady                         // Ctrl-I
	CodeEndBox                         // Ctrl-J
	CodeStartBox                       // Ctrl-K
	CodeNormalHeight                   // Ctrl-L
	CodeDoubleHeight                   // Ctrl-M
	_                                  //
	_                                  //
	CodeGraphicsBlack                  // Level 2.5 and above
	CodeGraphicsRed                    // Ctrl-F1
	CodeGraphicsGreen                  // Ctrl-F2
	CodeGraphicsYellow                 // Ctrl-F3
	CodeGraphicsBlue                   // Ctrl-F4
	CodeGraphicsMagenta                // Ctrl-F5
	CodeGraphicsCyan                   // Ctrl-F6
	CodeGraphicsWhite                  // Ctrl-F7
	CodeConcealDisplay                 // Ctrl-R
	CodeContiguousGraphics             // Ctrl-D (was Ctrl-Y)
	CodeSeparatedGraphics              // Ctrl-T
	CodeSwitch                         // ESC Toggles between the first and second G0 sets defined by packets X/28/0 Format 1, X/28/4, M/29/0 or M/29/4.
	CodeBlackBackground                // Ctrl-U
	CodeNewBackground                  // Ctrl-V
	CodeHoldGraphics                   // Ctrl-W
	CodeReleaseGraphics                // Ctrl-X
)
