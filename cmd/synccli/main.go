// Package main is the entry point for the sneakernet-sync CLI.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Aronwwo/sneakernet-sync/internal/core"
	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "sneakernet-sync",
		Short: "Offline file synchronization tool via external storage media",
	}

	root.AddCommand(
		newInitCmd(),
		newScanCmd(),
		newStatusCmd(),
		newPushCmd(),
		newPullCmd(),
		newSyncCmd(),
		newResolveCmd(),
		newConflictsCmd(),
		newDoctorCmd(),
	)

	return root
}

func newInitCmd() *cobra.Command {
	var deviceName string

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a sync repository in a directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			result, err := core.Init(dir, deviceName)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized sync repository\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Root:      %s\n", result.RootDir)
			fmt.Fprintf(cmd.OutOrStdout(), "  Meta:      %s\n", result.MetaDir)
			fmt.Fprintf(cmd.OutOrStdout(), "  Device ID: %s\n", result.DeviceID)
			return nil
		},
	}

	cmd.Flags().StringVar(&deviceName, "name", "", "Device name (defaults to hostname)")
	return cmd
}

func newScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan [directory]",
		Short: "Scan directory for changes and take a snapshot",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			result, err := engine.Scan()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Scan complete\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Files:       %d\n", result.Files)
			fmt.Fprintf(cmd.OutOrStdout(), "  Directories: %d\n", result.Directories)
			fmt.Fprintf(cmd.OutOrStdout(), "  Snapshot:    %s\n", result.SnapshotID)
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [directory]",
		Short: "Show current sync status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			status, err := engine.Status()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Sync Status\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  Device:        %s (%s)\n", status.DeviceName, status.DeviceID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Root:          %s\n", status.RootDir)
			fmt.Fprintf(cmd.OutOrStdout(), "  Last snapshot: %s\n", status.LastSnapshot)
			fmt.Fprintf(cmd.OutOrStdout(), "  Tracked files: %d\n", status.TrackedFiles)
			if status.HasConflicts {
				fmt.Fprintf(cmd.OutOrStdout(), "  Conflicts:     %d (unresolved)\n", status.ConflictCount)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  Conflicts:     none\n")
			}
			return nil
		},
	}
}

func newPushCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "push <media-path> [directory]",
		Short: "Export changes to external media",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mediaPath := args[0]
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			result, err := engine.Push(mediaPath, dryRun)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[DRY RUN] Push preview\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Push complete\n")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Snapshot: %s\n", result.SnapshotID)
			fmt.Fprintf(cmd.OutOrStdout(), "  Blobs:    %d\n", result.BlobCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Media:    %s\n", mediaPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newPullCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "pull <media-path> [directory]",
		Short: "Import changes from external media",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mediaPath := args[0]
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			result, err := engine.Pull(mediaPath, dryRun)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[DRY RUN] Pull preview\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Pull complete\n")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Remote device: %s\n", result.RemoteDevice)
			fmt.Fprintf(cmd.OutOrStdout(), "  Files:         %d\n", result.FileCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Blobs:         %d\n", result.BlobsFound)
			fmt.Fprintf(cmd.OutOrStdout(), "  Actions:       %d\n", result.Actions)
			if result.Conflicts > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  Conflicts:     %d\n", result.Conflicts)
			}

			for _, r := range result.Applied {
				if r.Error != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "  ! %s: %s\n", r.RelPath, r.Error)
				} else if r.OK {
					fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s [%s]\n", r.RelPath, r.Action)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newSyncCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync <media-path> [directory]",
		Short: "Full sync cycle (scan + push + pull)",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			mediaPath := args[0]
			dir := "."
			if len(args) > 1 {
				dir = args[1]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			// Step 1: Scan
			fmt.Fprintf(cmd.OutOrStdout(), "Step 1: Scanning...\n")
			scanResult, err := engine.Scan()
			if err != nil {
				return fmt.Errorf("scan: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Found %d files, %d directories\n", scanResult.Files, scanResult.Directories)

			// Step 2: Push
			fmt.Fprintf(cmd.OutOrStdout(), "Step 2: Pushing to media...\n")
			pushResult, err := engine.Push(mediaPath, dryRun)
			if err != nil {
				return fmt.Errorf("push: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  Exported %d blobs\n", pushResult.BlobCount)

			// Step 3: Pull (only if remote data exists on media)
			fmt.Fprintf(cmd.OutOrStdout(), "Step 3: Pulling from media...\n")
			pullResult, err := engine.Pull(mediaPath, dryRun)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  No remote data to pull: %v\n", err)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  Applied %d actions, %d conflicts\n", pullResult.Actions, pullResult.Conflicts)
			}

			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "\n[DRY RUN] No changes were written.\n")
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "\nSync complete.\n")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without writing")
	return cmd
}

func newResolveCmd() *cobra.Command {
	var resolution string

	cmd := &cobra.Command{
		Use:   "resolve <conflict-id>",
		Short: "Resolve a sync conflict",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conflictID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid conflict ID: %w", err)
			}

			engine, err := core.Open(".")
			if err != nil {
				return err
			}
			defer engine.Close()

			if err := engine.ResolveConflict(conflictID, resolution); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Conflict %d resolved with strategy: %s\n", conflictID, resolution)
			return nil
		},
	}

	cmd.Flags().StringVar(&resolution, "strategy", "manual", "Resolution strategy (keep-local, keep-remote, manual)")
	return cmd
}

func newConflictsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "conflicts [directory]",
		Short: "List unresolved sync conflicts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			conflicts, err := engine.GetConflicts()
			if err != nil {
				return err
			}

			if len(conflicts) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No unresolved conflicts.\n")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Unresolved conflicts (%d):\n\n", len(conflicts))
			for _, c := range conflicts {
				fmt.Fprintf(cmd.OutOrStdout(), "  ID:     %d\n", c.ID)
				fmt.Fprintf(cmd.OutOrStdout(), "  Path:   %s\n", c.RelPath)
				fmt.Fprintf(cmd.OutOrStdout(), "  Kind:   %s\n", c.Kind)
				fmt.Fprintf(cmd.OutOrStdout(), "  Local:  %s\n", c.LocalHash)
				fmt.Fprintf(cmd.OutOrStdout(), "  Remote: %s\n", c.RemoteHash)
				fmt.Fprintf(cmd.OutOrStdout(), "\n")
			}
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor [directory]",
		Short: "Verify repository integrity",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			engine, err := core.Open(dir)
			if err != nil {
				return err
			}
			defer engine.Close()

			issues, err := engine.Doctor()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Doctor report:\n")
			for _, issue := range issues {
				fmt.Fprintf(cmd.OutOrStdout(), "  • %s\n", issue)
			}
			return nil
		},
	}
}
