package xray

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCleanJSONC(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string // plain JSON the cleaned output must equal
	}{
		{"no comments", `{"a":1}`, `{"a":1}`},
		{"line comment", "{\n \"a\": 1 // one\n}", `{"a":1}`},
		{"line comment after comma", "{\n \"a\": 1, // one\n \"b\": 2\n}", `{"a":1,"b":2}`},
		{"block comment across lines", "{ /* x\n y */ \"a\": 1 }", `{"a":1}`},
		{"block comment splits tokens", `[1/**/,2]`, `[1,2]`},
		{"slashes inside a string", `{"u":"https://example.com/a//b"}`, `{"u":"https://example.com/a//b"}`},
		{"block marker inside a string", `{"s":"/* not a comment */"}`, `{"s":"/* not a comment */"}`},
		{"escaped quote then slashes", `{"s":"he said \"//no\" ok"}`, `{"s":"he said \"//no\" ok"}`},
		{"escaped backslash before quote", `{"s":"a\\","t":1} // x`, `{"s":"a\\","t":1}`},
		{"comment at EOF without newline", `{"a":1} // end`, `{"a":1}`},
		{"utf8 BOM", "\xef\xbb\xbf{\"a\":1}", `{"a":1}`},
		{"non-ascii comment", "{\"a\":1 // 换成你的域名\n}", `{"a":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanJSONC([]byte(tt.in))
			if err != nil {
				t.Fatalf("cleanJSONC() error = %v", err)
			}

			var g, w any
			if err := json.Unmarshal(got, &g); err != nil {
				t.Fatalf("cleaned output is not valid JSON: %v\n%q", err, got)
			}
			if err := json.Unmarshal([]byte(tt.want), &w); err != nil {
				t.Fatalf("bad test case: %v", err)
			}
			if !reflect.DeepEqual(g, w) {
				t.Fatalf("cleanJSONC() = %q, want JSON equal to %s", got, tt.want)
			}
		})
	}
}

func TestCleanJSONC_UnterminatedBlockComment(t *testing.T) {
	_, err := cleanJSONC([]byte(`{"a":1} /* oops`))
	if err == nil {
		t.Fatal("expected an error for an unterminated block comment")
	}
	if !strings.Contains(err.Error(), "unterminated") {
		t.Fatalf("error = %v, want unterminated message", err)
	}
}

func TestCleanJSONC_KeepsLineNumbers(t *testing.T) {
	in := "{\n// one\n/* two\nthree */\n\"a\": 1\n}"

	got, err := cleanJSONC([]byte(in))
	if err != nil {
		t.Fatalf("cleanJSONC() error = %v", err)
	}
	if strings.Count(string(got), "\n") != strings.Count(in, "\n") {
		t.Fatalf("newline count changed: got %d, want %d", strings.Count(string(got), "\n"), strings.Count(in, "\n"))
	}
}

// The example config from the Xray docs: comments after values, Chinese text,
// and a "//" that follows a comma.
func TestCleanJSONC_XrayDocsExample(t *testing.T) {
	in := `{
    "log": {
        "loglevel": "warning"
    },
    "outbounds": [
        {
            "protocol": "vless",
            "settings": {
                "address": "example.com", // 换成你的域名或服务器 IP（发起请求时无需解析域名了）
                "port": 443,
                "id": "", // 填写你的 UUID
                "encryption": "none",
                "level": 0
            },
            "streamSettings": {
                "network": "ws",
                "security": "tls",
                "tlsSettings": {
                    "serverName": "example.com" // 换成你的域名
                },
                "wsSettings": {
                    "path": "/websocket" // 必须换成自定义的 PATH，需要和服务端的一致
                }
            }
        }
    ]
}`

	// encoding/json must fail on the raw file...
	var raw any
	if err := json.Unmarshal([]byte(in), &raw); err == nil {
		t.Fatal("raw config with comments unexpectedly parsed")
	}

	// ...and succeed once cleaned, through the same path SaveOutboundFromFile uses.
	cleaned, err := cleanJSONC([]byte(in))
	if err != nil {
		t.Fatalf("cleanJSONC() error = %v", err)
	}

	var doc struct {
		Outbounds []json.RawMessage `json:"outbounds"`
	}
	if err := json.Unmarshal(cleaned, &doc); err != nil {
		t.Fatalf("cleaned config does not parse: %v", err)
	}
	if len(doc.Outbounds) != 1 {
		t.Fatalf("outbounds = %d, want 1", len(doc.Outbounds))
	}

	var o Outbound
	if err := json.Unmarshal(doc.Outbounds[0], &o); err != nil {
		t.Fatalf("outbound does not decode: %v", err)
	}
	if o.Protocol != ProtocolVLESS || o.Network() != NetworkWS || o.StreamSettings.TLSSettings.ServerName != "example.com" {
		t.Fatalf("outbound decoded wrongly: %#v", o)
	}
}
