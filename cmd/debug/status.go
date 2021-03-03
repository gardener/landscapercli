package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/cmd/debug/tree"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type statusOptions struct {
	kubeconfig       string
	installationName string
	namespace        string

	detailMode     bool
	showExecutions bool
	showOnlyFailed bool

	oyaml bool
	ojson bool
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

func NewStatusCommand(ctx context.Context) *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:     "status [installationName] [namespace] --kubeconfig [kubeconfig.yaml]",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(2),
		Example: "landscaper-cli debug status my-installation my-namespace --kubeconfig kubeconfig.yaml",
		Short:   "create an installation template for a component which is stored in an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArgs(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, cmd, logger.Log); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())
	cmd.MarkFlagRequired("kubeconfig")

	return cmd
}

func (o *statusOptions) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	// cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	// if err != nil {
	// 	return fmt.Errorf("cannot parse K8s config: %w", err)
	// }

	// k8sClient, err := client.New(cfg, client.Options{
	// 	Scheme: scheme,
	// })
	// if err != nil {
	// 	return fmt.Errorf("cannot build K8s client: %w", err)
	// }

	// coll := tree.Collector{
	// 	K8sClient: k8sClient,
	// }
	// installationTree, err := coll.CollectInstallationTree(o.installationName, o.namespace)
	// if err != nil {
	// 	return fmt.Errorf("cannot collect installation: %w", err)
	// }
	// transformedTree := tree.TransformToPrintableTree([]tree.InstallationTree{*installationTree})
	// output := tree.PrintTree(transformedTree)

	installationTrees := createDummyInstallationTree()

	if o.oyaml {
		marshaledInstallationTrees, err := yaml.Marshal(installationTrees)
		if err != nil {
			return fmt.Errorf("Failed marshaling output to yaml: %w", err)
		}
		cmd.Print(string(marshaledInstallationTrees))
		return nil
	}

	if o.ojson {
		marshaledInstallationTrees, err := json.Marshal(installationTrees)
		if err != nil {
			return fmt.Errorf("Failed marshaling output to json: %w", err)
		}
		cmd.Print(string(marshaledInstallationTrees))
		return nil
	}

	transformer := tree.TransformOptions{
		DetailedMode:   o.detailMode,
		ShowExecutions: o.showExecutions,
		OnlyFailed:     o.showOnlyFailed,
	}

	transformedTree, err := transformer.TransformToPrintableTree(installationTrees)
	if err != nil {
		cmd.PrintErrf("Error transforming CR to printable tree: %w", err)
	}
	output := tree.PrintTrees(transformedTree)

	cmd.Print(output.String())

	return nil
}

func (o *statusOptions) validateArgs(args []string) error {
	o.installationName = args[0]
	o.namespace = args[1]

	return nil
}

func (o *statusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig of the cluster")
	fs.BoolVarP(&o.detailMode, "details", "d", false, "show detailed information about installations, executions and deployitems")
	fs.BoolVarP(&o.showExecutions, "show-executions", "e", false, "show the executions in the tree")
	fs.BoolVarP(&o.showOnlyFailed, "show-failed", "f", false, "show failed items")
	fs.BoolVarP(&o.oyaml, "oyaml", "y", false, "output in yaml format")
	fs.BoolVarP(&o.ojson, "ojson", "j", false, "output in json format")
}

func createDummyInstallationTree() []*tree.InstallationTree {
	t := tree.InstallationTree{
		Installation: &lsv1alpha1.Installation{
			ObjectMeta: v1.ObjectMeta{Name: "main installation"},
			Status: lsv1alpha1.InstallationStatus{
				Phase: lsv1alpha1.ComponentPhaseFailed,
				LastError: &lsv1alpha1.Error{
					Message: "Long errorLong errorLong errorLong errorLong errorLong errorLong errorLong errorLong errorLong errorLong errorLong error",
				},
			},
		},
		Execution: &tree.ExecutionTree{
			Execution: &lsv1alpha1.Execution{
				ObjectMeta: v1.ObjectMeta{Name: "Execution"},
				Status: lsv1alpha1.ExecutionStatus{
					Phase: lsv1alpha1.ExecutionPhaseInit,
				},
			},
			DeployItems: []*tree.DeployItemLeaf{&tree.DeployItemLeaf{
				DeployItem: &lsv1alpha1.DeployItem{
					Status: lsv1alpha1.DeployItemStatus{
						Phase: lsv1alpha1.ExecutionPhaseFailed,
					},
					ObjectMeta: v1.ObjectMeta{Name: "depItem"},
				},
			},
				&tree.DeployItemLeaf{
					DeployItem: &lsv1alpha1.DeployItem{
						Status: lsv1alpha1.DeployItemStatus{
							Phase: lsv1alpha1.ExecutionPhaseFailed,
						},
						ObjectMeta: v1.ObjectMeta{Name: "depItem"},
					},
				}}},
		SubInstallations: []*tree.InstallationTree{
			&tree.InstallationTree{
				Installation: &lsv1alpha1.Installation{
					ObjectMeta: v1.ObjectMeta{Name: "first sub installation"},
					Status: lsv1alpha1.InstallationStatus{
						Phase: lsv1alpha1.ComponentPhaseSucceeded,
					},
				},

				Execution: &tree.ExecutionTree{
					Execution: &lsv1alpha1.Execution{
						ObjectMeta: v1.ObjectMeta{Name: "Execution"},
						Status: lsv1alpha1.ExecutionStatus{
							Phase: lsv1alpha1.ExecutionPhaseSucceeded,
						},
					},
					DeployItems: []*tree.DeployItemLeaf{&tree.DeployItemLeaf{
						DeployItem: &lsv1alpha1.DeployItem{
							Status: lsv1alpha1.DeployItemStatus{
								Phase:     lsv1alpha1.ExecutionPhaseSucceeded,
								LastError: &lsv1alpha1.Error{Message: "Error asdadasdasddasdadasdasd"},
							},
							ObjectMeta: v1.ObjectMeta{Name: "depItem"},
						},
					},
						&tree.DeployItemLeaf{
							DeployItem: &lsv1alpha1.DeployItem{
								Status: lsv1alpha1.DeployItemStatus{
									Phase: lsv1alpha1.ExecutionPhaseSucceeded,
								},
								ObjectMeta: v1.ObjectMeta{Name: "depItem"},
							},
						},
					},
				},
			},
			&tree.InstallationTree{
				Installation: &lsv1alpha1.Installation{
					ObjectMeta: v1.ObjectMeta{Name: "second sub installation"},
					Status: lsv1alpha1.InstallationStatus{
						Phase: lsv1alpha1.ComponentPhaseSucceeded,
					},
				},

				Execution: &tree.ExecutionTree{
					Execution: &lsv1alpha1.Execution{
						ObjectMeta: v1.ObjectMeta{Name: "Execution"},
						Status: lsv1alpha1.ExecutionStatus{
							Phase: lsv1alpha1.ExecutionPhaseInit,
						},
					},
					DeployItems: []*tree.DeployItemLeaf{&tree.DeployItemLeaf{
						DeployItem: &lsv1alpha1.DeployItem{
							Status: lsv1alpha1.DeployItemStatus{
								Phase:     lsv1alpha1.ExecutionPhaseFailed,
								LastError: &lsv1alpha1.Error{Message: "Error asdadasdasddasdadasdasd"},
							},
							ObjectMeta: v1.ObjectMeta{Name: "depItem"},
						},
					},
						&tree.DeployItemLeaf{
							DeployItem: &lsv1alpha1.DeployItem{
								Status: lsv1alpha1.DeployItemStatus{
									Phase: lsv1alpha1.ExecutionPhaseSucceeded,
								},
								ObjectMeta: v1.ObjectMeta{Name: "depItem Not failed"},
							},
						},
					},
				},
			},
		},
	}
	return []*tree.InstallationTree{&t, &t}
}
