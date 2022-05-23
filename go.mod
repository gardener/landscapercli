module github.com/gardener/landscapercli

go 1.16

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.5.9
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/mandelsoft/spiff => github.com/mandelsoft/spiff v1.3.0-beta-7.0.20200909122641-3393af1d3804
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.8.0
	golang.org/x/text => golang.org/x/text v0.3.5
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/ahmetb/gen-crd-api-reference-docs v0.3.0
	github.com/gardener/component-cli v0.36.0
	github.com/gardener/component-spec/bindings-go v0.0.53
	github.com/gardener/landscaper v0.24.0
	github.com/gardener/landscaper/apis v0.24.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/mock v1.5.0
	github.com/mandelsoft/vfs v0.0.0-20210530103237-5249dc39ce91
	github.com/onsi/ginkgo v1.16.4
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.19.0
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
	sigs.k8s.io/controller-runtime v0.10.0
	sigs.k8s.io/yaml v1.2.0
)
