module github.com/gardener/landscapercli

go 1.15

require (
	github.com/gardener/component-spec/bindings-go v0.0.0-20201214104938-0bdb01b553e8
	github.com/gardener/landscaper v0.0.0-20201210105611-436fc88acc40
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/mandelsoft/vfs v0.0.0-20201002134249-3c471f64a4d1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.15.0
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/mandelsoft/spiff => github.com/mandelsoft/spiff v1.3.0-beta-7.0.20200909122641-3393af1d3804
