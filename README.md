# k8s-reset-terminating-pv

Reset persistent volume status from terminating back to bound. [Here are the details](https://jianz.github.io/posts/2020-08-25-reset-pv).

## Purpose

When delete a kubernetes persistent volume by accident, it may stuck in the terminating status due to `kubernetes.io/pv-protection` finalizer prevent it from being deleted. You can use this tool to reset its status back to bound.

Now you can change pvc status too.

## Installing

You can download the latest compiled binary from [here](https://github.com/jianz/k8s-reset-terminating-pv/releases).

If you prefer to compile by yourself:

```shell
git clone git@github.com:jianz/k8s-reset-terminating-pv.git
cd k8s-reset-terminating-pv
go build -o resetpv
```

## Usage

```text
Usage:
  resetpv [flags] <resource resourceName [namespace]>

Flags:
      --etcd-ca        string   CA Certificate used by etcd (default "ca.crt")
      --etcd-cert      string   Public key used by etcd (default "etcd.crt")
      --etcd-key       string   Private key used by etcd (default "etcd.key")
      --etcd-host      string   The etcd domain name or IP (default "localhost")
      --etcd-port      int      The etcd port number (default 2379)
      --k8s-key-prefix string   The etcd key prefix for kubernetes resources. (default "registry")
  -h, --help                    help for resetpv
```

For simplicity, you can name the etcd certificate ca.crt, etcd.crt, etcd.key, and put them in the same directory as the tool(resetpv).

The tool by default connect to etcd using `localhost:2379`. You can forward the etcd port on the pod to the localhost:

```shell
kubectl port-forward pods/etcd-member-master0 2379:2379 -n etcd
```

`--k8s-key-prefix`: Default set to `registry` for the community version of kubernetes as it uses `/registry` as etcd key prefix, the key for persistent volume pv1 is `/registry/persistentvolumes/pv1`. Set to `kubernetes.io` for OpenShift as it uses `/kubernetes.io` as prefix and the key for pv1 is `/kubernetes.io/persistentvolumes/pv1`.

Example:

```shell
./resetpv --k8s-key-prefix kubernetes.io pv pv-eef4ec4b-326d-47e6-b11c-6474a5fd4d89
```

if you want to change pvc status do as follow
```shell
./resetpv  pvc pvcname namespace
```
## License

k8s-reset-terminating-pv is released under the MIT license.
