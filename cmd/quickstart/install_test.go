package quickstart

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
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
			if err := tt.opts.Complete([]string{}); tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestCheckConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		opts        installOptions
		expectedErr string
	}{
		{
			name: "install OCI registry with ingress (correctly configured)",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryIngressHost: "registry.ingress.my-cluster.com",
				landscaperValues: landscaperValues{
					Landscaper: landscaperconfig{
						Landscaper: landscaper{
							RegistryConfig: registryConfig{
								AllowPlainHttpRegistries: false,
								Secrets: secrets{
									Defaults: defaults{
										Auths: map[string]interface{}{
											"some-other-registry-url.com": map[string]interface{}{
												"auth": "dummy",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "install OCI registry without ingress (correctly configured)",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: false,
				landscaperValues: landscaperValues{
					Landscaper: landscaperconfig{
						Landscaper: landscaper{
							RegistryConfig: registryConfig{
								AllowPlainHttpRegistries: true,
								Secrets: secrets{
									Defaults: defaults{
										Auths: map[string]interface{}{},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "install OCI registry without ingress and allowPlainHttpRegistries = false",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: false,
				landscaperValues: landscaperValues{
					Landscaper: landscaperconfig{
						Landscaper: landscaper{
							RegistryConfig: registryConfig{
								AllowPlainHttpRegistries: false,
								Secrets: secrets{
									Defaults: defaults{
										Auths: map[string]interface{}{},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "landscaper.landscaper.registryConfig.allowPlainHttpRegistries must be set to true when installing Landscaper together with the OCI registry without ingress access",
		},
		{
			name: "install OCI registry with ingress and allowPlainHttpRegistries = true",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryIngressHost: "registry.ingress.my-cluster.com",
				landscaperValues: landscaperValues{
					Landscaper: landscaperconfig{
						Landscaper: landscaper{
							RegistryConfig: registryConfig{
								AllowPlainHttpRegistries: true,
								Secrets: secrets{
									Defaults: defaults{
										Auths: map[string]interface{}{},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "landscaper.landscaper.registryConfig.allowPlainHttpRegistries must be set to false when installing Landscaper together with the OCI registry with ingress access",
		},
		{
			name: "include registry credentials already in user defined Landscaper values",
			opts: installOptions{
				instOCIRegistry:     true,
				instRegistryIngress: true,
				registryIngressHost: "registry.ingress.my-cluster.com",
				landscaperValues: landscaperValues{
					Landscaper: landscaperconfig{
						Landscaper: landscaper{
							RegistryConfig: registryConfig{
								AllowPlainHttpRegistries: false,
								Secrets: secrets{
									Defaults: defaults{
										Auths: map[string]interface{}{
											"registry.ingress.my-cluster.com": map[string]interface{}{
												"auth": "dummy",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "registry credentials for registry.ingress.my-cluster.com are already set in the provided Landscaper values. please remove them from your configuration as they will get configured automatically",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.opts.checkConfiguration(); tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestGenerateLandscaperValuesOverride(t *testing.T) {
	opts := installOptions{
		instOCIRegistry:     true,
		instRegistryIngress: false,
		landscaperValues: landscaperValues{
			Landscaper: landscaperconfig{
				Landscaper: landscaper{
					RegistryConfig: registryConfig{
						AllowPlainHttpRegistries: true,
						Secrets: secrets{
							Defaults: defaults{
								Auths: map[string]interface{}{},
							},
						},
					},
				},
			},
		},
	}

	// verify that default deployers are added if not specified
	lsvo, err := opts.generateLandscaperValuesOverride()
	assert.NoError(t, err)
	data := map[string]interface{}{}
	err = yaml.Unmarshal(lsvo, &data)
	assert.NoError(t, err)
	config, ok := data["landscaper"]
	assert.True(t, ok)
	data, ok = config.(map[string]interface{})
	assert.True(t, ok)
	config, ok = data["landscaper"]
	assert.True(t, ok)
	data, ok = config.(map[string]interface{})
	assert.True(t, ok)
	config, ok = data["deployers"]
	assert.True(t, ok)
	depList, ok := config.([]interface{})
	assert.True(t, ok)
	assert.EqualValues(t, []interface{}{"container", "helm", "manifest"}, depList)

	// verify that deployers are not overwritten if specified
	opts.landscaperValues.Landscaper.Landscaper.Deployers = []string{"mock"}
	lsvo, err = opts.generateLandscaperValuesOverride()
	assert.NoError(t, err)
	data = map[string]interface{}{}
	err = yaml.Unmarshal(lsvo, &data)
	assert.NoError(t, err)
	config, ok = data["landscaper"]
	assert.True(t, ok)
	data, ok = config.(map[string]interface{})
	assert.True(t, ok)
	config, ok = data["landscaper"]
	assert.True(t, ok)
	data, ok = config.(map[string]interface{})
	assert.True(t, ok)
	_, ok = data["deployers"]
	assert.False(t, ok)
}
