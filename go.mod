module bpxe.org

go 1.16

require (
	github.com/antchfx/jsonquery v1.1.4
	github.com/antchfx/xpath v1.1.11
	github.com/antonmedv/expr v1.8.9
	github.com/go-enry/go-license-detector/v4 v4.2.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/mitchellh/go-homedir v1.1.0
	// Latest sno doesn't exhibit SSE2 bug (https://github.com/muyo/sno/issues/3)
	github.com/muyo/sno v1.1.1-0.20200406142550-e5d36f06b5d6
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
)
