/*
Package xbin implements the eXtended BIN (BinaryText) Format specification.


Introduction

Out of a crying need from several ANSi artists, a new type of file is born and
ready to take the Art scene by storm.

This new format is known as eXtended BIN or XBIN for short.

XBIN is what it's name dictates, it's an extention to the normal raw-image BIN
files which have become very popular among the art-scene lately.

The use of the XBIN format is more or less the same as for the BIN format.
However, XBIN offers a far better way to handle the the raw images.


BIN vs XBIN

The BIN format was introduced into the art-scene out of a need to overcome
the limits of ANSi files.  Apparantly, the 80 columns wide screen was too
constraining for some artists.  The BIN format was adopted to resolve this
problem.

Being very simple in nature, BIN was quickly supported by several art
groups in their native ANSI/RIP/GIF/etc. viewer.  Consequently, our very
own SAUCE standard went in for a quick facelift and immediately dealt with
one of the main problems imposed by the BIN format.

Being nothing more than a raw memory copy of the textmode videomemory, BIN
offers no insight to the size/width of the image.  Having nothing more than
a BIN file, there is no way to determine whether it's a 80 column wide or
a 160 column wide image.  SAUCE took care of this by taking the BIN format
into it's specifications.  Out of the SAUCE attached to the BIN, one was
able to determine the correct dimensions of the BIN.

XBIN solves this little matter all by itself, and takes matters even
further.  Anything BIN can do, XBIN does better.
*/
package xbin
