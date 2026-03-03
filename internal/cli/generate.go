package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trianalab/pacto/internal/app"
)

func newGenerateCommand(svc *app.Service, v *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate <plugin> [path | oci://ref]",
		Short: "Generate artifacts from a contract using a plugin",
		Long:  "Invokes a pacto-plugin-<name> binary to generate deployment manifests, documentation, or other artifacts from a contract.",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]
			var path string
			if len(args) > 1 {
				path = args[1]
			}

			outputDir, _ := cmd.Flags().GetString("output")

			result, err := svc.Generate(cmd.Context(), app.GenerateOptions{
				Path:      path,
				OutputDir: outputDir,
				Plugin:    pluginName,
			})
			if err != nil {
				return err
			}

			format := v.GetString("output-format")
			return printGenerateResult(cmd, result, format)
		},
	}

	cmd.Flags().StringP("output", "o", "", "output directory (default: <plugin>-output/)")

	return cmd
}
