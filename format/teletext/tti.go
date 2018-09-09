package teletext

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// DecodeTTI decodes a TTI encoded set of pages.
func DecodeTTI(r io.Reader) (Pages, error) {
	var (
		b        = bufio.NewReader(r)
		line     string
		op, args string
		lineno   int
		pages    Pages
		page     = NewPage()
		root     = page
		i64      int64  // helper
		u64      uint64 // helper
		err      error
	)

	pages = append(pages, root)

parsing:
	for {
		lineno++
		if line, err = b.ReadString('\n'); err != nil {
			if err == io.EOF {
				break parsing
			}
		}

		line = strings.TrimRight(line, "\r\n")
		if i := strings.IndexByte(line, ','); i != 2 {
			if i == -1 {
				return nil, fmt.Errorf("teletext: line %d: no op found", lineno)
			}
			return nil, fmt.Errorf("teletext: line %d: op of length %d found", lineno, i+1)
		}

		switch op, args = line[:2], line[3:]; op {
		case "DE": // Description
			page.description = args

		case "DS": // Destination inserter name
			page.destination = args

		case "PN": // Page number
			// Format: mppss
			//    m  = 1..8
			//    pp = 00 to ff (hex)
			//    ss = 00 to 99 (decimal)
			if len(args) < 3 {
				return nil, fmt.Errorf("teletext: line %d: illegal page number %q: EOF", lineno, args)
			} else if args[0] < '1' || args[0] > '8' {
				return nil, fmt.Errorf("teletext: line %d: illegal page number %q: invalid magazine %q", lineno, args, args[0])
			}

			var pp uint64
			if pp, err = strconv.ParseUint(args[1:], 16, 32); err != nil {
				return nil, fmt.Errorf("teletext: line %d: illegal page number %q: %v", lineno, args, err)
			}
			if len(args) < 5 && pp <= 0x8ff {
				pp *= 0x100
			}

			if page.Number != FirstPage {
				status, lang, cycleTime, CycleTimeType := page.status, page.Language, page.CycleTime, page.CycleTimeType
				page = NewPage()
				page.status = status
				page.Language = lang
				page.CycleTime = cycleTime
				page.CycleTimeType = CycleTimeType
				pages = append(pages, page)
			}
			page.Number = int(pp)

		case "PS": // Page status flags
			// Format: nnn
			//   nnn = 000 to fff (hex)
			if i64, err = strconv.ParseInt(args, 16, 32); err != nil {
				return nil, fmt.Errorf("teletext: line %d: illegal page status %q: %v", lineno, args, err)
			}
			page.status = int(i64)

		case "CT": // Cycle Time
			// Format: cc OR cc,t
			//   cc = 00 to 99 (decimal seconds)
			//   t  = C or T   (optional)
			if i, l := strings.IndexByte(args, ','), len(args); i != -1 && i < l {
				if args[i+1] == 'T' {
					page.CycleTimeType = 'T'
				} else {
					page.CycleTimeType = 'C'
				}
				args = args[:i]
			}
			if page.CycleTime, err = strconv.Atoi(args); err != nil {
				return nil, fmt.Errorf("teletext: line %d: illegal cycle time: %v", lineno, err)
			}

		case "RE": // Set page region code 0..f
			if len(args) != 1 {
				return nil, fmt.Errorf("teletext: line %d: illegal region code %q", lineno, args)
			}
			page.region = int(args[0] - '0')

		case "SC": // Subcode
			if u64, err = strconv.ParseUint(args, 16, 32); err != nil {
				return nil, fmt.Errorf("teletext: line %d: illegal subcode %q: %v", lineno, args, err)
			}
			page.subCode = uint(u64)

		case "OL": // Output Line
			// OL,9,A-Z INDEX     199NEWS HEADLINES  101
			var i int
			if i = strings.IndexByte(args, ','); i == -1 {
				return nil, fmt.Errorf("teletext: line %d: illegal output line %q", lineno, args)
			}
			var row int
			if row, err = strconv.Atoi(args[:i]); err != nil {
				return nil, fmt.Errorf("teletext: line %d: illegal output line %q: %v", lineno, args, err)
			}
			page.SetRow(uint8(row), []byte(args[i+1:]))

		case "FL": // Fastext links
			// FL,104,104,105,106,F,100
			for i, link := range strings.Split(args, ",") {
				if u64, err = strconv.ParseUint(link, 16, 32); err != nil {
					return nil, fmt.Errorf("teletext: line %d: fastext link line %q: %v", lineno, args, err)
				}
				page.FastExtLinks[i] = int(u64)
			}

		case "PF": // Page function and coding
			if len(args) != 3 {
				return nil, fmt.Errorf("teletext: line %d: page function line %q", lineno, args)
			}
			if u64, err = strconv.ParseUint(args[:1], 16, 32); err != nil {
				return nil, fmt.Errorf("teletext: line %d: page function line %q: %v", lineno, args, err)
			}
			page.function = PageFunction(u64)
			if u64, err = strconv.ParseUint(args[2:], 16, 32); err != nil {
				return nil, fmt.Errorf("teletext: line %d: page function line %q: %v", lineno, args, err)
			}
			page.coding = PageCoding(u64)

		case "MS": // Mask
		case "RD":
		case "SP": // Source page file name
			// ignored

		default:
			return nil, fmt.Errorf("teletext: line %d: illegal op %q", lineno, op)
		}
	}

	return pages, nil
}
