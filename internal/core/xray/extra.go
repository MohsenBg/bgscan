package xray

import (
	"bytes"
	"encoding/json"
	"errors"
)

// extraFields returns the top-level keys of the JSON object in data that are
// not listed in known. Typed structs use it to keep fields they do not model,
// so decoding and re-encoding a template never silently drops settings.
func extraFields(data []byte, known map[string]struct{}) (map[string]json.RawMessage, error) {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return nil, err
	}

	var extra map[string]json.RawMessage
	for k, v := range all {
		if _, ok := known[k]; ok {
			continue
		}
		if extra == nil {
			extra = make(map[string]json.RawMessage)
		}
		extra[k] = v
	}

	return extra, nil
}

// marshalWithExtra encodes v and merges extra into the resulting object.
// v must be a method-free alias of the typed struct (to avoid recursing into
// its own MarshalJSON). Typed fields win if a key exists in both.
func marshalWithExtra(v any, extra map[string]json.RawMessage) ([]byte, error) {
	base, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(extra) == 0 {
		return base, nil
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(base, &m); err != nil {
		return nil, err
	}
	for k, raw := range extra {
		if _, typed := m[k]; !typed {
			m[k] = raw
		}
	}

	return json.Marshal(m)
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// cleanJSONC turns an Xray-style config (JSON with // and /* */ comments,
// optionally starting with a UTF-8 BOM) into plain JSON that encoding/json
// accepts. Comment markers inside strings are left alone, and newlines inside
// comments are kept so parse errors still point at the right line.
func cleanJSONC(src []byte) ([]byte, error) {
	src = bytes.TrimPrefix(src, utf8BOM)
	out := make([]byte, 0, len(src))
	inString := false

	for i := 0; i < len(src); i++ {
		c := src[i]

		if inString {
			out = append(out, c)
			switch c {
			case '\\': // keep the escaped byte, even if it is a quote
				if i+1 < len(src) {
					i++
					out = append(out, src[i])
				}
			case '"':
				inString = false
			}
			continue
		}

		switch {
		case c == '"':
			inString = true
			out = append(out, c)

		case c == '/' && i+1 < len(src) && src[i+1] == '/':
			for i < len(src) && src[i] != '\n' {
				i++
			}
			if i < len(src) {
				out = append(out, '\n')
			}

		case c == '/' && i+1 < len(src) && src[i+1] == '*':
			i += 2
			for i+1 < len(src) && (src[i] != '*' || src[i+1] != '/') {
				if src[i] == '\n' {
					out = append(out, '\n')
				}
				i++
			}
			if i+1 >= len(src) {
				return nil, errors.New("unterminated /* comment")
			}
			i++ // now on the closing '/'
			out = append(out, ' ')

		default:
			out = append(out, c)
		}
	}

	return out, nil
}
