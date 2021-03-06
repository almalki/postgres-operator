---
title: "pgo load"
---
## pgo load

Perform a data load

### Synopsis

LOAD performs a load. For example:

	pgo load --load-config=./load.json --selector=project=xray

```
pgo load [flags]
```

### Options

```
  -h, --help                 help for load
      --load-config string   The load configuration to use that defines the load job.
      --policies string      The policies to apply before loading a file, comma separated.
  -s, --selector string      The selector to use for cluster filtering.
```

### Options inherited from parent commands

```
      --apiserver-url string     The URL for the PostgreSQL Operator apiserver.
      --debug                    Enable debugging when true.
      --disable-tls              Disable TLS authentication to the Postgres Operator.
      --exclude-os-trust         Exclude CA certs from OS default trust store
  -n, --namespace string         The namespace to use for pgo requests.
      --pgo-ca-cert string       The CA Certificate file path for authenticating to the PostgreSQL Operator apiserver.
      --pgo-client-cert string   The Client Certificate file path for authenticating to the PostgreSQL Operator apiserver.
      --pgo-client-key string    The Client Key file path for authenticating to the PostgreSQL Operator apiserver.
```

### SEE ALSO

* [pgo](/operatorcli/cli/pgo/)	 - The pgo command line interface.

###### Auto generated by spf13/cobra on 23-Dec-2019
