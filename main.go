package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/hpcloud/tail"
	"gopkg.in/alexcesaro/statsd.v2"
)

// Parse nginx log until http referer.
// More info here: http://nginx.org/en/docs/http/ngx_http_log_module.html
var r = regexp.MustCompile(`^(?P<remote>[^ ]*) (?P<host>[^ ]*) (?P<user>[^ ]*) \[(?P<time>[^\]]*)\] "(?P<method>\S+)(?: +(?P<path>[^ ]*) +\S*)?" (?P<code>[^ ]*) (?P<size>[^ ]*)`)

func ParseNginxLog(line string) (map[string]string, error) {
	match := r.FindStringSubmatch(line)
	data := make(map[string]string)

	if len(r.SubexpNames()) != len(match) {
		return data, errors.New("Could not parse the line:\n" + line)
	}

	for i, name := range r.SubexpNames() {
		if i != 0 {
			data[name] = match[i]
		}
	}

	data["path_normalized"] = NormalizePath(data["path"])

	return data, nil
}

type c struct {
	regex   string
	replace string
}

func NormalizePath(path string) string {
	cleanUp := []c{
		c{
			regex:   "^/js/.*",
			replace: "js",
		},
		c{
			regex:   "^/css/.*",
			replace: "css",
		},
		c{
			regex:   "^/img/.*",
			replace: "img",
		},
		c{
			regex:   "^/.*\\.php",
			replace: "php",
		},

		c{
			regex:   "[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}",
			replace: "uuid",
		},
		c{
			regex:   "[0-9a-fA-F]{24}",
			replace: "mongoid",
		},
		c{
			regex:   "\\?.*$",
			replace: "/query",
		},
		c{
			regex:   "/+",
			replace: "/",
		},
		c{
			regex:   "^/",
			replace: "",
		},
		c{
			regex:   "/$",
			replace: "",
		},
		c{
			regex:   "/",
			replace: ".",
		},
	}

	var result = strings.ToLower(path)

	for _, item := range cleanUp {
		var r = regexp.MustCompile(item.regex)
		result = r.ReplaceAllString(result, item.replace)
	}

	return result
}

func verifyFile(fn string) {
	stopAt := 300
	normalizedData := make(map[string]int)

	file, err := os.Open(fn)
	defer file.Close()

	if err != nil {
		fmt.Printf("Failed opening file: %v\n", err)
		return
	}

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var line string
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Failed reading line: %v\n", err)
			break
		}

		if strings.TrimSpace(line) == "" {
			break
		}

		stats, err := ParseNginxLog(line)
		if err != nil {
			fmt.Printf("Failed to parse line: %v\n", err)
			break
		}

		normalized := stats["path_normalized"]
		normalizedData[normalized]++

		if len(normalizedData) >= stopAt {
			fmt.Printf("Size > %v\n", stopAt)
			break
		}
	}

	fmt.Printf("normalizedData:\n")
	for k, v := range normalizedData {
		fmt.Printf("%v: %v\n", k, v)
	}
	fmt.Printf("%v of %v \n", len(normalizedData), stopAt)
}

func main() {
	var host = flag.String("host", "localhost", "StatsD host")
	var port = flag.Int("port", 8125, "StatsD port")
	var prefix = flag.String("prefix", "nginx.access.log", "StatsD metrics prefix")
	var file = flag.String("file", "/var/log/nginx/access.log", "Nginx log file")

	flag.Parse()
	
	address := fmt.Sprintf("%s:%d", *host, *port)
	stats, err := statsd.New(statsd.Address(address))
	if err != nil {
		log.Panic(err)
	}
	defer stats.Close()

	info := tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	config := tail.Config{
		Location: &info,
		Follow:   true,
		ReOpen:   true,
	}

	t, err := tail.TailFile(*file, config)
	if err != nil {
		log.Panic(err)
	}

	for line := range t.Lines {
		data, err := ParseNginxLog(line.Text)
		if err != nil {
			log.Println(err)
			stats.Increment(fmt.Sprintf("%s.%s", *prefix, "unknown"))
		}
		stats.Increment(fmt.Sprintf("%s.%s", *prefix, data["code"]))
	}
}
