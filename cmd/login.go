package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/internal/auth"
	"github.com/suer/first-task-cli/internal/config"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in with your Google account",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			cred, err := auth.Login(cmd.Context(), cfg, openBrowser)
			if err != nil {
				return err
			}
			fmt.Printf("Logged in as %s (uid: %s)\n", cred.Email, cred.UID)
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.DeleteCredentials()
		},
	}
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the logged-in user",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cred, err := config.LoadCredentials()
			if err != nil {
				return err
			}
			fmt.Printf("%s (uid: %s)\n", cred.Email, cred.UID)
			return nil
		},
	}
}
