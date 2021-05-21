module github.com/gardener/landscapercli

go 1.16

replace (
	github.com/gardener/landscaper/apis => github.com/gardener/landscaper/apis v0.8.3
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	github.com/mandelsoft/spiff => github.com/mandelsoft/spiff v1.3.0-beta-7.0.20200909122641-3393af1d3804
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.8.0
	golang.org/x/text => golang.org/x/text v0.3.5
)

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/ahmetb/gen-crd-api-reference-docs v0.2.0
	github.com/gardener/component-cli v0.19.0
	github.com/gardener/component-spec/bindings-go v0.0.39
	github.com/gardener/landscaper v0.8.3
	github.com/gardener/landscaper/apis v0.8.3
	github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr v0.3.0
	github.com/golang/mock v1.4.4
	github.com/mandelsoft/vfs v0.0.0-20201002134249-3c471f64a4d1
	github.com/onsi/ginkgo v1.14.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)
