package sauce

// Data types for the Record.DataType field
const (
	None uint8 = iota
	Character
	Bitmap
	Vector
	Audio
	BinaryText
	XBIN
	FileType
	Executable
)

// File types for the Record.FileType field
const (
	// Character
	ASCII = iota
	ANSi
	ANSiMation
	RIPscript
	PCBoard
	Avatar
	HTML
	Source
	TundraDraw
)

// Bitmap file types
const (
	GIF  = iota
	PCX  // ZSoft Paintbrush PCX
	LBM  // DeluxePaint LBM/IFF
	TGA  // Targa Truecolor
	FLI  // Autodesk FLI animation
	FLC  // Autodesk FLC animation
	BMP  // Windows or OS/2 Bitmap
	GL   // Grasp GL Animation
	DL   // DL Animation
	WPG  // Wordperfect Bitmap
	PNG  // Portable Network Graphics
	JPEG //	JPEG image (any subformat)
	MPEG // MPEG video (any subformat)
	AVI  // Audio Video Interleave (any subformat)
)

// Vector file types
const (
	DXF        = iota // CAD Drawing eXchange Format
	DWG               // AutoCAD Drawing File
	WPGraphics        // WordPerfect or DrawPerfect vector graphics
	ThreeDS           // 3D Studio
)

// Audio file types
const (
	MOD          = iota // 4, 6 or 8 channel MOD (NoiseTracker)
	Renaissance8        // Renaissance 8 channel 669
	STM                 // Future Crew 4 channel ScreamTracker
	S3M                 // Future Crew variable channel ScreamTracker 3
	MTM                 // Renaissance variable channel MultiTracker
	FAR                 // Farandole composer
	ULT                 // UltraTracker
	AMF                 // DMP/DSMI Advanced Module Format
	DMF                 // Delusion Digital Music Format (XTracker)
	OKT                 // Oktalyser
	ROL                 // AdLib ROL file (FM audio)
	CMF                 // Creative Music File (FM Audio)
	MID                 // MIDI (Musical Instrument Digital Interface)
	SADT                //	SAdT composer (FM Audio)
	VOC                 // Creative Voice File
	WAV                 // Waveform Audio File Format
	SMP8                // Raw, single channel 8bit sample	Sample rate [7]
	SMP8S               // Raw, stereo 8 bit sample	Sample rate [7]
	SMP16               // Raw, single-channel 16 bit sample	Sample rate [7]
	SMP16S              // Raw, stereo 16 bit sample	Sample rate [7]
	PATCH8              // 8 Bit patch file
	PATCH16             // 16 bit patch file
	XM                  // FastTracker ][ module
	HSC                 // HSC Tracker (FM Audio)
	IT                  // Impulse Tracker
)

// Archive file types
const (
	ZIP = iota // PKWare Zip.
	ARJ        // Archive Robert K. Jung
	LZH        // Haruyasu Yoshizaki (Yoshi)
	ARC        // S.E.A.
	TAR        // Unix TAR
	ZOO        // ZOO
	RAR        // RAR
	UC2        // UC2
	PAK        // PAK
	SQZ        // SQZ
)
