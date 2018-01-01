package main

import (
	"testing"
)

func TestParseNginxLog(t *testing.T) {
	var log = `127.0.0.1 - - [01/Jan/2018:17:23:43 +0100] "GET /index.html HTTP/1.1" 200 612 "-" "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.89 Safari/537.36"`

	data, err := ParseNginxLog(log)

	if err != nil {
		t.Error("Could not parse:\n", log)
	}

	if data["code"] != "200" {
		t.Error("Expected", "200", "got", data["code"])
	}
}
