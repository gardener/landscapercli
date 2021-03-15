package quickstart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplete(t *testing.T) {
	tests := []struct {
		name        string
		opts        installOptions
		expectedErr string
	}{
		{
			name: "don't install OCI registry",
			opts: installOptions{
				instOCIRegistry:     false,
				instRegistryIngress: false,
				registryUsername:    "",
				registryPassword:    "",
			},
		},
		{
			name: "install OCI registry without ingress",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: false,
				registryUsername:    "",
				registryPassword:    "",
			},
		},
		{
			name: "install OCI registry and ingress with credentials set",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryUsername:    "user",
				registryPassword:    "123456",
			},
		},
		{
			name: "install OCI registry and ingress with missing username",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryUsername:    "",
				registryPassword:    "123456",
			},
			expectedErr: "username must be provided if --install-registry-ingress is set to true",
		},
		{
			name: "install OCI registry and ingress with missing password",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryUsername:    "user",
				registryPassword:    "",
			},
			expectedErr: "password must be provided if --install-registry-ingress is set to true",
		},
		{
			name: "invalid configuration 1",
			opts: installOptions{
				instOCIRegistry:     false,
				instRegistryIngress: true,
				registryUsername:    "user",
				registryPassword:    "123456",
			},
			expectedErr: "you can only set --install-registry-ingress to true together with --install-oci-registry",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Complete([]string{})
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestCheckConfiguration(t *testing.T) {
	
}
