module github.com/jianz/k8s-reset-terminating-pv

go 1.15

require (
	github.com/coreos/etcd v3.3.24+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/spf13/cobra v1.0.0
	go.etcd.io/etcd v3.3.24+incompatible
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/grpc v1.31.0 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
