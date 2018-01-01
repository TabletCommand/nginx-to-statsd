package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

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

	return data, nil
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

	seek_info := tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	config := tail.Config{
		Location: &seek_info,
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
