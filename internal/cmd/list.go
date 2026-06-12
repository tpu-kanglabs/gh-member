package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	ghterm "github.com/cli/go-gh/v2/pkg/term"
	"github.com/spf13/cobra"
	xterm "golang.org/x/term" // for terminal width

	"github.com/tpu-kanglabs/gh-member/internal/gh"
	"github.com/tpu-kanglabs/gh-member/internal/output"
	"github.com/tpu-kanglabs/gh-member/internal/ui"
)

// newClientFn is the factory used to create the GraphQL client.
// Overridden in tests to inject a fake client without a live GitHub session.
var newClientFn func() (gh.GraphQLDoer, error) = gh.NewDefaultClient

func newListCmd() *cobra.Command {
	var (
		limit    int
		role     string
		jsonFlag string
		jqFilter string
		tmplStr  string
	)

	cmd := &cobra.Command{
		Use:   "list [org]",
		Short: "List members of a GitHub Organization",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var org string
			if len(args) > 0 {
				org = args[0]
			}

			// validate --role
			switch strings.ToLower(role) {
			case "all", "admin", "member":
				// valid
			default:
				return fmt.Errorf("invalid role %q: must be one of all, admin, member", role)
			}

			// structured output mode?
			structuredMode := jsonFlag != "" || jqFilter != "" || tmplStr != ""

			// Validate org early (before creating client) for deterministic error messages.
			t := ghterm.FromEnv()
			isTTY := t.IsTerminalOutput()
			if structuredMode && org == "" {
				return errors.New("organization is required when using --json, --jq, or --template")
			}
			if !structuredMode && !isTTY && org == "" {
				return errors.New("organization is required in non-interactive mode")
			}

			client, err := newClientFn()
			if err != nil {
				return fmt.Errorf("create GraphQL client: %w", err)
			}

			ctx := context.Background()

			if structuredMode {
				members, err := gh.FetchMembers(ctx, client, org, limit, role)
				if err != nil {
					return err
				}

				var fields []string
				if jsonFlag != "" {
					fields = strings.Split(jsonFlag, ",")
				}

				var buf bytes.Buffer
				if err := output.PrintJSON(&buf, members, fields); err != nil {
					return err
				}

				stdout := cmd.OutOrStdout()

				switch {
				case jqFilter != "":
					return output.ApplyJQ(stdout, &buf, jqFilter)
				case tmplStr != "":
					return output.ApplyTemplate(stdout, &buf, tmplStr)
				default:
					_, err = io.Copy(stdout, &buf)
					return err
				}
			}

			// non-structured mode
			if isTTY {
				// TUI mode
				if org == "" {
					orgs, err := gh.FetchViewerOrgs(ctx, client, 100)
					if err != nil {
						return err
					}
					selected, err := ui.SelectOrg(orgs)
					if err != nil {
						return err
					}
					if selected == "" {
						// user cancelled
						return nil
					}
					org = selected
				}

				members, err := gh.FetchMembers(ctx, client, org, limit, role)
				if err != nil {
					return err
				}

				return ui.BrowseMembers(members)
			}

			// static table mode (non-TTY)
			members, err := gh.FetchMembers(ctx, client, org, limit, role)
			if err != nil {
				return err
			}

			return output.PrintTable(t.Out(), isTTY, termWidth(), members)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "L", 30, "Maximum number of members to fetch (-1 for all)")
	cmd.Flags().StringVar(&role, "role", "all", "Filter by role: all, admin, member")
	cmd.Flags().StringVar(&jsonFlag, "json", "", "Output fields as JSON (comma-separated field names)")
	cmd.Flags().StringVarP(&jqFilter, "jq", "q", "", "Apply jq filter to JSON output")
	cmd.Flags().StringVarP(&tmplStr, "template", "t", "", "Apply Go template to JSON output")

	return cmd
}

// termWidth returns the current terminal width, defaulting to 80.
func termWidth() int {
	width, _, err := xterm.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}
