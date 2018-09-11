/*
Package bgi implements Borland Graphics Interface images.
*/
package bgi

/*
The CHR files are scalable fonts used by the Borland graphics interface
(BGI) to display fonts in graphics mode.
OFFSET              Count TYPE   Description
0000h                   4 char   ID='PK',08,08
0004h                   4 char   ID='BGI '
0008h                   ? char   Font description, terminated with #26
0008h                   1 word   Headersize
+????                            ="SIZ"
						4 char   Internal font name
						1 word   Font file size in bytes
						1 byte   Font driver major version
						1 byte   Font driver minor version
						1 word   0100h
					"SIZ" word   Zeroes to pad out the header
0080h                   1 char   Signature byte, '+' means stroke font
0081h                   1 word   Number of chars in font file
								 ="NUM"
0083h                   1 byte   undefined
0084h                   1 byte   ASCII value of first char in file
0085h                   1 word   Offset to stroke definitions
0087h                   1 byte   Scan flag ??
0088h                   1 byte   Distance from origin to top of capital
0089h                   1 byte   Distance from origin to baseline
008Ah                   1 byte   Distance from origin to bottom descender
008Bh                   4 char   Four character name of font
0090h               "NUM" word   Offsets to character definitions
0090h+              "NUM" byte   Width table for the characters
"NUM"*2
0090h+                           Start of character definitions
"NUM"*3

The individual character definitions consist of a variable number of words
describing the operations required to render a character. Each word
consists of an (x,y) coordinate pair and a two-bit opcode, encoded as shown
here:

Byte 1          7   6   5   4   3   2   1   0     bit #
			   op1  <seven bit signed X coord>

Byte 2          7   6   5   4   3   2   1   0     bit #
			   op2  <seven bit signed Y coord>

		  Opcodes

		op1=0  op2=0  End of character definition.
		op1=0  op2=1  Do scan
		op1=1  op2=0  Move the pointer to (x,y)
		op1=1  op2=1  Draw from current pointer to (x,y)
*/
