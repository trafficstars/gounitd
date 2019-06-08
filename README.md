Please look at [envoyproxy](https://www.envoyproxy.io/) first.

```sh
go get https://github.com/trafficstars/gounitd
go install https://github.com/trafficstars/gounitd
cp $(go env GOPATH)/src/github.com/trafficstars/gounitd/gounit.yaml.sample /etc/gounit.yaml
vim /etc/gounit.yaml
$(go env GOPATH)/bin/gounitd -config /etc/gounit.yaml
```
