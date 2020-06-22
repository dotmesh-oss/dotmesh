package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/dotmesh-oss/dotmesh/pkg/client"
	"github.com/spf13/cobra"
)

func NewCmdInit(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <dot>",
		Short: "Create an empty dot",
		Long:  "Online help: https://docs.dotmesh.com/references/cli/#create-an-empty-dot-dm-init-dot",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				dm, err := client.NewDotmeshAPI(configPath, verboseOutput)
				if err != nil {
					return err
				}
				if len(args) > 1 {
					return fmt.Errorf("Too many arguments specified (more than 1).")
				}
				if len(args) == 0 {
					return fmt.Errorf("No dot name specified.")
				}
				v := args[0]
				if !client.CheckName(v) {
					return fmt.Errorf("Error: %v is an invalid name: dot names must be <50 characters", v)
				}
				exists, err := dm.VolumeExists(v)
				if err != nil {
					return err
				}
				if exists {
					return fmt.Errorf("Error: %v exists already", v)
				}
				err = dm.NewVolume(v)
				if err != nil {
					return fmt.Errorf("Error: %v", err)
				}
				return nil
			}()
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		},
	}
	return cmd
}
