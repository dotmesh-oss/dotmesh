package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/dotmesh-oss/dotmesh/pkg/client"
	"github.com/spf13/cobra"
)

func NewCmdSwitch(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch",
		Short: "Change which dot is active",
		Long:  "Online help: https://docs.dotmesh.com/references/cli/#select-a-different-current-dot-dm-switch-dot",
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
				volumeName := args[0]
				exists, err := dm.VolumeExists(volumeName)
				if err != nil {
					return err
				}
				if !exists {
					return fmt.Errorf("Error: %v doesn't exist", volumeName)
				}
				err = dm.SwitchVolume(volumeName)
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
