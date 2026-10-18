package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/internal/auth"
	"github.com/suer/first-task-cli/internal/config"
	"github.com/suer/first-task-cli/internal/firestore"
)

var jsonOutput bool

func Execute() error {
	root := &cobra.Command{
		Use:           "firsttask",
		Short:         "Manage FirstTask data on Firestore",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(newLoginCmd(), newLogoutCmd(), newWhoamiCmd(), newTaskCmd(), newProjectCmd(), newTagCmd())

	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return err
	}
	return nil
}

func newClient(ctx context.Context) (*firestore.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	cred, err := auth.Session(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return firestore.New(cfg.ProjectID, cred.UID, cred.IDToken), nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func openBrowser(url string) {
	fmt.Fprintf(os.Stderr, "Open this URL in your browser to log in:\n%s\n", url)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}
