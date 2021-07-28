module github.com/champly/lib4go

go 1.15

require (
	github.com/champly/gojenkins v1.0.0
	github.com/docker/docker v20.10.1+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-zookeeper/zk v1.0.2
	github.com/gogo/protobuf v1.3.2
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/mailru/easyjson v0.7.2 // indirect
	github.com/olivere/elastic v6.2.34+incompatible
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/prometheus/client_golang v1.7.1
	github.com/tealeg/xlsx v1.0.5
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200910180754-dd1b699fc489
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5 // indirect
	golang.org/x/text v0.3.4
	google.golang.org/grpc v1.31.0
	helm.sh/helm/v3 v3.6.3
	istio.io/client-go v1.8.1
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/klog/v2 v2.8.0
	mosn.io/pkg v0.0.0-20200729115159-2bd74f20be0f
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)
