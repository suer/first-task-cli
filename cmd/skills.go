package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Songmu/skillsmith"
	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/skills"
)

type skillsFlags struct {
	scope  string
	prefix string
	dryRun bool
	force  bool
}

func (f *skillsFlags) options() skillsmith.Options {
	return skillsmith.Options{Scope: f.scope, Prefix: f.prefix, DryRun: f.dryRun, Force: f.force}
}

func (f *skillsFlags) addTo(cmd *cobra.Command, withModify bool) {
	cmd.Flags().StringVar(&f.scope, "scope", "", "install scope (user or repo)")
	cmd.Flags().StringVar(&f.prefix, "prefix", "", "override install directory")
	if withModify {
		cmd.Flags().BoolVar(&f.dryRun, "dry-run", false, "preview changes without applying")
		cmd.Flags().BoolVar(&f.force, "force", false, "overwrite unmanaged skills or force downgrade")
	}
}

func newSmith() (*skillsmith.Smith, error) {
	v := currentVersion()
	if v == "dev" {
		v = "v0.0.0-dev"
	}
	return skillsmith.New("firsttask", v, skills.FS)
}

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "Manage agent skills",
	}
	cmd.AddCommand(
		newSkillsListCmd(),
		newSkillsModifyCmd("install", "Install skills", "installed", (*skillsmith.Smith).Install),
		newSkillsModifyCmd("update", "Update installed skills", "updated", (*skillsmith.Smith).Update),
		newSkillsModifyCmd("reinstall", "Reinstall all managed skills", "reinstalled", (*skillsmith.Smith).Reinstall),
		newSkillsUninstallCmd(),
		newSkillsStatusCmd(),
	)
	return cmd
}

func newSkillsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSmith()
			if err != nil {
				return err
			}
			list, err := s.List(cmd.Context())
			if err != nil {
				return err
			}
			if len(list) == 0 {
				fmt.Println("no skills found")
				return nil
			}
			for _, sk := range list {
				fmt.Printf("%-20s %s\n", sk.Dir, sk.Description)
			}
			return nil
		},
	}
}

func newSkillsModifyCmd(use, short, verb string, run func(*skillsmith.Smith, context.Context, skillsmith.Options) (*skillsmith.CopyResult, error)) *cobra.Command {
	var f skillsFlags
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSmith()
			if err != nil {
				return err
			}
			opts := f.options()
			result, err := run(s, cmd.Context(), opts)
			if err != nil {
				return err
			}
			for _, a := range result.Actions {
				switch a.Action {
				case verb:
					fmt.Printf("%s%s: %s\n", verb, dryRunSuffix(opts.DryRun), a.Dir)
				case "skipped":
					fmt.Printf("skipped: %s - %s\n", a.Dir, a.Message)
				case "warned":
					fmt.Fprintf(os.Stderr, "warning: %s - %s\n", a.Dir, a.Message)
				}
			}
			if opts.DryRun {
				fmt.Println("[dry-run] no changes were made")
			}
			return nil
		},
	}
	f.addTo(cmd, true)
	return cmd
}

func newSkillsUninstallCmd() *cobra.Command {
	var f skillsFlags
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall managed skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSmith()
			if err != nil {
				return err
			}
			opts := f.options()
			result, err := s.Uninstall(cmd.Context(), opts)
			if err != nil {
				return err
			}
			for _, a := range result.Actions {
				switch a.Action {
				case "uninstalled":
					fmt.Printf("uninstalled%s: %s\n", dryRunSuffix(opts.DryRun), a.Dir)
				case "skipped":
					fmt.Printf("skipped: %s - %s\n", a.Dir, a.Message)
				}
			}
			if opts.DryRun {
				fmt.Println("[dry-run] no changes were made")
			}
			return nil
		},
	}
	f.addTo(cmd, true)
	return cmd
}

func newSkillsStatusCmd() *cobra.Command {
	var f skillsFlags
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show installation status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newSmith()
			if err != nil {
				return err
			}
			result, err := s.Status(cmd.Context(), f.options())
			if err != nil {
				return err
			}
			for _, ss := range result.Skills {
				switch {
				case !ss.Installed:
					fmt.Printf("%-20s not installed\n", ss.Dir)
				case ss.MetadataError != nil:
					fmt.Printf("%-20s installed (metadata unreadable: %v)\n", ss.Dir, ss.MetadataError)
				case ss.UpToDate:
					fmt.Printf("%-20s installed %s (up to date)\n", ss.Dir, ss.InstalledVersion)
				default:
					fmt.Printf("%-20s installed %s -> available %s\n", ss.Dir, ss.InstalledVersion, ss.AvailableVersion)
				}
			}
			return nil
		},
	}
	f.addTo(cmd, false)
	return cmd
}

func dryRunSuffix(dryRun bool) string {
	if dryRun {
		return " (dry-run)"
	}
	return ""
}
