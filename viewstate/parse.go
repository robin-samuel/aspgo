package viewstate

import (
	"errors"
	"fmt"
	"image/color"
	"time"
)

func parse(body []byte) (any, []byte, error) {
	if len(body) < 2 {
		return nil, nil, errors.New("invalid viewstate, too short")
	}

	marker, remain := body[0], body[1:]
	switch marker {
	case 0x01:
		return nil, remain, nil
	case 0x64:
		return nil, remain, nil
	case 0x65:
		return "", remain, nil
	case 0x66:
		return 0, remain, nil
	case 0x67:
		return true, remain, nil
	case 0x68:
		return false, remain, nil
	case 0x02, 0x2B:
		return parseInteger(remain)
	case 0x05, 0x1E, 0x2A, 0x29:
		return parseString(remain)
	case 0x0B:
		return parseEnum(remain)
	case 0x0A:
		return parseColor(remain)
	case 0x0F:
		return parsePair(remain)
	case 0x10:
		return parseTriplet(remain)
	case 0x06:
		return parseTime(remain)
	case 0x1B:
		return parseUnit(remain)
	case 0x09:
		return parseRGBA(remain)
	case 0x15:
		return parseStringSlice(remain)
	case 0x16:
		return parseSlice(remain)
	case 0x1F:
		return parseStringRef(remain)
	case 0x28:
		return parseFormattedString(remain)
	case 0x3C:
		return parseSparseArray(remain)
	case 0x18:
		return parseMap(remain)
	case 0x14:
		return parseTypedSlice(remain)
	case 0x32:
		return parseBinary(remain)
	}

	return nil, nil, errors.New("invalid viewstate, unknown marker")
}

func parseInteger(body []byte) (int, []byte, error) {
	var n, bits, i int
	for bits < 32 {
		tmp := body[i]
		i++
		n |= (int(tmp) & 0x7f) << bits
		if tmp&0x80 == 0 {
			break
		}
		bits += 7
	}
	return n, body[i:], nil
}

func parseString(body []byte) (string, []byte, error) {
	n, body, err := parseInteger(body)
	if err != nil {
		return "", nil, err
	}
	return string(body[:n]), body[n:], nil
}

func parseEnum(body []byte) (string, []byte, error) {
	enum, body, err := parse(body)
	if err != nil {
		return "", nil, err
	}
	val, body, err := parseInteger(body)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%v:%d", enum, val), body, nil
}

func parseColor(body []byte) (string, []byte, error) {
	return fmt.Sprintf("color:%v", body[:1]), body[1:], nil
}

func parsePair(body []byte) (any, []byte, error) {
	first, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	second, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	return [2]any{first, second}, body, nil
}

func parseTriplet(body []byte) (any, []byte, error) {
	first, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	second, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	third, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	return [3]any{first, second, third}, body, nil
}

func parseTime(body []byte) (time.Time, []byte, error) {
	// TODO
	return time.Time{}, body[8:], nil
}

func parseUnit(body []byte) (any, []byte, error) {
	// TODO
	return "unit", body[12:], nil
}

func parseRGBA(body []byte) (color.RGBA, []byte, error) {
	color := color.RGBA{
		R: body[0],
		G: body[1],
		B: body[2],
		A: body[3],
	}
	return color, body[4:], nil
}

func parseStringSlice(body []byte) ([]string, []byte, error) {
	n, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	slice := make([]string, n)
	for i := 0; i < n; i++ {
		slice[i], body, err = parseString(body)
		if err != nil {
			return nil, nil, err
		}
	}
	return slice, body, nil
}

func parseSlice(body []byte) ([]any, []byte, error) {
	n, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	slice := make([]any, n)
	for i := 0; i < n; i++ {
		slice[i], body, err = parse(body)
		if err != nil {
			return nil, nil, err
		}
	}
	return slice, body, nil
}

func parseStringRef(body []byte) (string, []byte, error) {
	val, body, err := parseInteger(body)
	return fmt.Sprintf("ref:%d", val), body, err
}

func parseFormattedString(body []byte) (string, []byte, error) {
	s1, body, err := parse(body)
	if err != nil {
		return "", nil, err
	}
	s2, body, err := parseString(body)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%v %v", s1, s2), body, nil
}

func parseSparseArray(body []byte) ([]any, []byte, error) {
	typ, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	_ = typ
	length, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	n, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	lst := make([]any, length)
	for i := 0; i < n; i++ {
		var idx int
		var val any
		idx, body, err = parseInteger(body)
		if err != nil {
			return nil, nil, err
		}
		val, body, err = parse(body)
		if err != nil {
			return nil, nil, err
		}
		lst[idx] = val
	}
	return lst, body, nil
}

func parseMap(body []byte) (map[string]any, []byte, error) {
	n := int(body[0])
	body = body[1:]
	m := make(map[string]any, n)
	for i := 0; i < n; i++ {
		var key, val any
		var err error
		key, body, err = parse(body)
		if err != nil {
			return nil, nil, err
		}
		val, body, err = parse(body)
		if err != nil {
			return nil, nil, err
		}
		m[fmt.Sprintf("%v", key)] = val
	}
	return m, body, nil
}

func parseTypedSlice(body []byte) ([]any, []byte, error) {
	typ, body, err := parse(body)
	if err != nil {
		return nil, nil, err
	}
	_ = typ
	n, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	slice := make([]any, n)
	for i := 0; i < n; i++ {
		slice[i], body, err = parse(body)
		if err != nil {
			return nil, nil, err
		}
	}
	return slice, body, nil
}

func parseBinary(body []byte) ([]byte, []byte, error) {
	n, body, err := parseInteger(body)
	if err != nil {
		return nil, nil, err
	}
	return body[:n], body[n:], nil
}
