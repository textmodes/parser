/*
Package sauce implements the Standard Architecture for Universal Comment
Extensions (SAUCE) standard.

The Standard Architecture for Universal Comment Extensions or SAUCE as it is
more commonly known, is an architecture or protocol for attaching meta data or
comments to files. Mainly intended for ANSI art files, SAUCE has always had
provisions for many different file types.


Why SAUCE

In the early 1990s there was a growing popularity in ANSI artwork. The ANSI art
groups regularly released the works of their members over a certain period.
Some of the bigger groups also included specialised viewers in each ‘artpack’.

One of the problems with these artpacks was a lack of standardized way to
provide meta data to the art, such as the title of the artwork, the author, the
group, ... Some of the specialised viewers provided such information for a
specific artpack either by encoding it as part of the executable, or by having
some sort of database or list. However every viewer did it their own way. This
meant you either had to use the viewer included with the artpack, or had to
make do without the extra info. SAUCE was created to address that need. So if
you wanted to, you could use your prefered viewer to view the art in a certain
artpack, or even store the art files you liked in a separate folder while
retaining the meta data.

The goal was simple, but the way to get there certainly was not. Logistically,
we wanted as many art groups as possible to support it. Technically, we wanted
a system that was easy to implement and – if at all possible – manage to
provide this meta data while still being compatible with all the existing
software such as ANSI viewers, and Bulletin Board Software.


The SAUCE Legacy

SAUCE was created in 1994 and it continues to be in use today as a de facto
standard within the ANSI art community. However, being created so many years
ago, some of the common assumptions made back then are now cause for confusion.

The ‘born from DOS’ nature explains some of the limits of SAUCE. The Title,
Author and Group fields are so short because part of the original design idea
was that the DOS filename, title and author needed to fit on a single line in
text mode while still leaving some space so you could create a decent UI to
select the files in the ANSI viewers. Limitations were also designed around
what video cards could do in text mode.

The specialised viewers in the various ANSI artpacks also explain why it was
possible to add SAUCE to some files even though the file format really does not
like it when you just add a load of extra bytes at the end. The SAUCE-aware
viewer could account for the SAUCE and still render the file properly, even if
the applications used to edit those files regarded the file as corrupt. In such
an event, it was easy enough to remove the SAUCE.

SAUCE is not a perfect solution, but at the time – and with a bit of friendly
pressure from the right folks – it managed to fulfill what it was designed to
do.
*/
package sauce
