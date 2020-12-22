module github.com/champly/lib4go

go 1.15

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/champly/gojenkins v1.0.0
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/fortytw2/leaktest v1.3.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-zookeeper/zk v1.0.2
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.4.3
	github.com/googleapis/gnostic v0.5.3 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mailru/easyjson v0.7.2 // indirect
	github.com/olivere/elastic v6.2.34+incompatible
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.11.0
	github.com/tealeg/xlsx v1.0.5
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5 // indirect
	golang.org/x/text v0.3.4
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	istio.io/client-go v1.8.1
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.19.2
	k8s.io/klog/v2 v2.4.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	mosn.io/pkg v0.0.0-20200729115159-2bd74f20be0f
	sigs.k8s.io/controller-runtime v0.7.0
)

replace (
	k8s.io/api => k8s.io/api v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.19.0
)
