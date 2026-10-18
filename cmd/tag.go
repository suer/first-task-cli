package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/suer/first-task-cli/internal/model"
)

func newTagCmd() *cobra.Command {
	tag := &cobra.Command{Use: "tag", Short: "Manage tags"}

	tag.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List tags",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			tags, err := loadTags(cmd.Context(), c)
			if err != nil {
				return err
			}
			if jsonOutput {
				return printJSON(tags)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tKIND")
			for _, t := range tags {
				fmt.Fprintf(w, "%s\t%s\t%s\n", shortID(t.ID), t.Name, t.Kind)
			}
			return w.Flush()
		},
	})

	tag.AddCommand(&cobra.Command{
		Use:   "add NAME",
		Short: "Create a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			t := model.Tag{ID: newID(), Name: args[0]}
			if err := c.Create(cmd.Context(), model.TagsCollection, t.Document()); err != nil {
				return err
			}
			fmt.Println(t.ID)
			return nil
		},
	})

	tag.AddCommand(&cobra.Command{
		Use:   "rm ID",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := newClient(cmd.Context())
			if err != nil {
				return err
			}
			tags, err := loadTags(cmd.Context(), c)
			if err != nil {
				return err
			}
			ids := make([]string, len(tags))
			for i, t := range tags {
				ids[i] = t.ID
			}
			i, err := matchID(ids, args[0])
			if err != nil {
				return err
			}
			return c.Delete(cmd.Context(), model.TagsCollection, tags[i].ID)
		},
	})

	return tag
}
