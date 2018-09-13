def mosaic(x):
    m = [' ', ' ', ' ', ' ', ' ', ' ']
    for i in range(6):
        if x & 1<<i:
            m[i] = '#'

    print('\n'.join([
        m[0] + m[1],
        m[2] + m[3],
        m[4] + m[5],
    ]))


def mosaic(x):
    """
    0b00001111
    0b00001111
    0b00001111
    0b11110000
    0b11110000
    0b11110000
    0b00001111
    0b00001111
    0b00001111
    """
    o = bytearray(9)
    if x & 1:
        o[0] |= 0xf0
        o[1] |= 0xf0
        o[2] |= 0xf0
    if x & 2:
        o[0] |= 0x0f
        o[1] |= 0x0f
        o[2] |= 0x0f
    if x & 4:
        o[3] |= 0xf0
        o[4] |= 0xf0
        o[5] |= 0xf0
    if x & 8:
        o[3] |= 0x0f
        o[4] |= 0x0f
        o[5] |= 0x0f
    if x & 16:
        o[6] |= 0xf0
        o[7] |= 0xf0
        o[8] |= 0xf0
    if x & 32:
        o[6] |= 0x0f
        o[7] |= 0x0f
        o[8] |= 0x0f

    return o

with open('mosaic.bin', 'wb') as handle:
    for x in range(64):
        handle.write(mosaic(x))

for x in range(64):
    print(x, mosaic(x))
