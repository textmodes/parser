#!/usr/bin/env python

import argparse
import base64
import sys


def convert(source, output=sys.stdout):
    if '#' in source:
        # Assume it's a full URL, just get the hash
        source = source.split('#', 2)[1]
    if not source:
        raise ValueError('Source is empty string')

    i = source.index(':')
    data, code = source[i+1:], source[:i]
    data = decode7bits(data)

    output.write(data)
    if output.isatty():
        output.write('\n')


def decode7bits(data):
    #if len(data) == 1120 or len(data) == 1167:
    #    #return decode7bitbase64(data)

    if len(data) == 1120:
        data = base64.urlsafe_b64decode(data)
    elif len(data) == 1167:
        data = base64.urlsafe_b64decode(data+'_')
    else:
        raise ValueError("Unknown length {}".format(len(data)))

    output = bytearray()
    while len(data) > 6:
        output.extend(bytearray([
            (ord(data[0]) >> 1) & 0x7f,
            (ord(data[0]) << 6) & 0x40 | (ord(data[1]) >> 2) & 0x3f,
            (ord(data[1]) << 5) & 0x60 | (ord(data[2]) >> 3) & 0x1f,
            (ord(data[2]) << 4) & 0x70 | (ord(data[3]) >> 4) & 0x0f,
            (ord(data[3]) << 3) & 0x78 | (ord(data[4]) >> 5) & 0x07,
            (ord(data[4]) << 2) & 0x7c | (ord(data[5]) >> 6) & 0x03,
            (ord(data[5]) << 1) & 0x7e | (ord(data[6]) >> 7) & 0x01,
            (ord(data[6]) << 0) & 0x7f,
        ]))
        data = data[7:]

    return output


BASE64_ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"


def decode7bitbase64(data):
    # As we scan across the hashstring, we keep track of the code for the
    # current character cell we're writing into.
    code = 0

    output = bytearray()
    for p, c in enumerate(data):
        # p is the position in the data
        # c is the encoded value
        # d is the decoded index
        d = BASE64_ALPHABET.index(c)

        # b is the bit in the 6-bit base-64
        b = 0
        while b < 6:
            chrbit = (6 * p + b) % 7
            b64bit = d & (1 << (5 - b))
            if b64bit > 0:
                b64bit = 1

            code |= b64bit << (6 - chrbit)
            if chrbit == 6:
                chrnum = ((6 * p + b) - chrbit) / 7
                n = (chrnum    ) % 40
                r = (chrnum - n) / 40
                #if placeable(code) == 1:
                output.append(code)
                code = 0

            b += 1

    return output



def placeable(code):
    return ( code >= 0  and code <= 7 ) \
            or ( code == 8 or code == 9 ) \
            or ( code == 12 or code == 13 ) \
            or ( code >= 16 and code <= 23 ) \
            or ( code == 24 ) \
            or ( code == 25 or code == 26 ) \
            or ( code == 28 or code == 29 ) \
            or ( code == 30 or code == 31 )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('-o', '--output', default='', type=str)
    parser.add_argument('input', type=str)
    args = parser.parse_args()

    output = sys.stdout
    if args.output:
        output = open(args.output, 'wb')

    try:
        convert(args.input, output)
    finally:
        output.close()


if __name__ == '__main__':
    sys.exit(main())
