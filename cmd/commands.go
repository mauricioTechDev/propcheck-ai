package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Show all available commands and flags",
	Long:  "Dump the entire CLI reference for AI agents to learn the full API in one call.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		type flagInfo struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Default string `json:"default,omitempty"`
			Usage   string `json:"usage"`
		}
		type cmdInfo struct {
			Name     string     `json:"name"`
			Short    string     `json:"short"`
			Usage    string     `json:"usage"`
			Flags    []flagInfo `json:"flags,omitempty"`
			Children []cmdInfo  `json:"children,omitempty"`
		}

		var collectCmd func(c *cobra.Command) cmdInfo
		collectCmd = func(c *cobra.Command) cmdInfo {
			info := cmdInfo{
				Name:  c.Name(),
				Short: c.Short,
				Usage: c.UseLine(),
			}
			c.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if f.Hidden {
					return
				}
				info.Flags = append(info.Flags, flagInfo{
					Name:    f.Name,
					Type:    f.Value.Type(),
					Default: f.DefValue,
					Usage:   f.Usage,
				})
			})
			for _, sub := range c.Commands() {
				if sub.Hidden || sub.Name() == "help" || sub.Name() == "completion" {
					continue
				}
				info.Children = append(info.Children, collectCmd(sub))
			}
			return info
		}

		root := collectCmd(rootCmd)

		if formatFlag == "json" {
			data, err := json.MarshalIndent(root, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		var printCmd func(c cmdInfo, indent int)
		printCmd = func(c cmdInfo, indent int) {
			prefix := strings.Repeat("  ", indent)
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s — %s\n", prefix, c.Usage, c.Short)
			for _, f := range c.Flags {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  --%s (%s) %s\n", prefix, f.Name, f.Type, f.Usage)
			}
			for _, sub := range c.Children {
				printCmd(sub, indent+1)
			}
		}
		printCmd(root, 0)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}
