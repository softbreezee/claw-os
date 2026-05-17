package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/softbreezee/claw-os/internal/mcpserver"
)

// mcpReasonixCmd starts an MCP stdio server that exposes Reasonix-style
// diagnostic and repair tools. Pawnix agents connect to it by adding
// an mcpServers entry in their config:
//
//	"mcpServers": {
//	  "reasonix": {
//	    "type": "stdio",
//	    "command": "pawnix",
//	    "args": ["mcp-reasonix"]
//	  }
//	}
//
// Tools provided: check_config, check_skills, check_disk, analyze_logs,
// repair_config, search_files, search_content, list_directory.
func mcpReasonixCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp-reasonix",
		Short: "Start the Reasonix MCP server (diagnostic/repair tools)",
		Long: `Start a Reasonix-branded MCP (Model Context Protocol) server on stdio.

Pawnix agents can connect to this server via their mcpServers config
entry and use the diagnostic/repair tools to self-heal the platform.

Tools:
  check_config   — validate pawnix.json integrity
  check_skills   — validate all installed skill files
  check_disk     — report disk usage and thresholds
  analyze_logs   — scan logs for error/warn patterns
  repair_config  — fix common config issues
  search_files   — find files by name pattern
  search_content — grep through file contents
  list_directory — list directory with sizes and dates

Use "pawnix start" for the normal gateway; this command only runs the
MCP server in the foreground for the agent client to attach.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			srv := mcpserver.NewServer("Reasonix", "0.1.0")
			mcpserver.RegisterReasonixTools(srv)
			if err := srv.Serve(); err != nil {
				fmt.Fprintf(os.Stderr, "reasonix mcp server: %v\n", err)
				return err
			}
			return nil
		},
	}
}
