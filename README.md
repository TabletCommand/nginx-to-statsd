# nginx-to-statsd
Send Nginx response code metrics to statsd

```
docker build --tag nginx-to-statsd .
docker run -d --network="host" --volume="/var/log:/var/log" nginx-to-statsd
```

Usage of nginx-to-statsd:
* file `string` Nginx log file (default "/var/log/nginx/access.log")
* host `string` StatsD host (default "localhost")
* port `int` StatsD port (default 8125)
* prefix `string` StatsD metrics prefix (default "nginx.access.log")
