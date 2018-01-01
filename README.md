# nginx-to-statsd
Send Nginx response code metrics to statsd

```
docker build --tag nginx-to-statsd .
docker run -d --network="host" --volume="/var/log:/var/log" nginx-to-statsd
```
