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

func TestNormalizePath(t *testing.T) {
	data := map[string]string{
		"/api/sync/items/A54E6F3D-2850-434C-AFDA-D138627A89AE": "api.sync.items.uuid",
		"/api/sync/items/?modified_unix_date=15":               "api.sync.items.query",
		"/api/sync/items/592ba375e47dac145f680bab":             "api.sync.items.mongoid",
		"/js/external/ie10-viewport-bug-workaround.js":         "js",
		"/css/file.min.css":                                    "css",
		"/img/image.png":                                       "img",
		"/system.php":                                          "php",
	}

	for input, output := range data {
		result := NormalizePath(input)
		if result != output {
			t.Error("Expected", output, "got", result, "for", input)
		}
	}
}
