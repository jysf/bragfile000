package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/jysf/bragfile000/internal/config"
	"github.com/jysf/bragfile000/internal/export"
	"github.com/jysf/bragfile000/internal/storage"
	"github.com/spf13/cobra"
)

// getCwd is the function called to get the current working directory.
// Package-level so tests in the cli package can override it without
// the production binary carrying any test-only surface.
var getCwd = os.Getwd

// NewProjectCmd returns the `brag project` parent command with new,
// list, show, status, edit, archive, delete, and here subcommands. A bare
// `brag project` prints help (cobra default for a command with subcommands
// and no RunE).
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects (new, list, show, status, edit, archive, delete, here)",
	}
	cmd.AddCommand(newProjectNewCmd())
	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectStatusCmd())
	cmd.AddCommand(newProjectEditCmd())
	cmd.AddCommand(newProjectArchiveCmd())
	cmd.AddCommand(newProjectDeleteCmd())
	cmd.AddCommand(newProjectHereCmd())
	return cmd
}

func newProjectNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <name> --path <dir>",
		Short: "Register a new project with an initial location",
		Long: `Register a new project with a filesystem location. The project starts with
status "active" and an empty state note; use 'brag project edit' to change them.
The --path is required and is stored verbatim; a path already registered to
another project is rejected.

Examples:
  brag project new bragfile --path ~/code/bragfile
  brag project new platform --path /srv/platform`,
		RunE: runProjectNew,
	}
	cmd.Flags().String("path", "", "filesystem directory for the project (required)")
	return cmd
}

func runProjectNew(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("new requires exactly one <name> argument")
	}
	name := strings.TrimSpace(args[0])
	if name == "" {
		return UserErrorf("project name must not be empty")
	}
	path, _ := cmd.Flags().GetString("path")
	if strings.TrimSpace(path) == "" {
		return UserErrorf("--path is required (the project's directory)")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	// LD3: pre-check the path is free so a conflict creates no orphan
	// project. ListProjects hydrates Locations; iterating its result is
	// plain Go (no SQL in the CLI layer). Exact-string match — the same
	// verbatim basis AddLocation enforces; SPEC-031 owns normalization.
	existing, err := s.ListProjects()
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}
	for _, p := range existing {
		for _, loc := range p.Locations {
			if loc == path {
				return UserErrorf("path %q is already registered to project %q", path, p.Name)
			}
		}
	}

	created, err := s.CreateProject(storage.Project{Name: name})
	if err != nil {
		if errors.Is(err, storage.ErrProjectExists) {
			return UserErrorf("project %q already exists", name)
		}
		return fmt.Errorf("create project: %w", err)
	}
	if err := s.AddLocation(created.ID, path); err != nil {
		if errors.Is(err, storage.ErrLocationExists) {
			// Defensive backstop for the TOCTOU window (no real race in a
			// single-user CLI); the pre-check above is the primary guard.
			return UserErrorf("path %q is already registered to another project", path)
		}
		return fmt.Errorf("add location: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Created project %q.\n", name)
	return nil
}

func newProjectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects (most-recently-updated first)",
		Long: `List every registered project, most-recently-updated first, one per line as
<name>` + "\t" + `<status>` + "\t" + `<locations> (comma-joined; "-" when none).

Output is plain tab-separated rows (default) or a naked JSON array of project
objects (--format json) per DEC-011.

Examples:
  brag project list                 # name<TAB>status<TAB>locations
  brag project list --format json   # naked JSON array of project objects`,
		RunE: runProjectList,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain tab-separated")
	return cmd
}

func runProjectList(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	projects, err := s.ListProjects()
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectsJSON(projects)
		if err != nil {
			return fmt.Errorf("render projects json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	for _, p := range projects {
		loc := "-"
		if len(p.Locations) > 0 {
			loc = strings.Join(p.Locations, ",")
		}
		fmt.Fprintf(out, "%s\t%s\t%s\n", p.Name, p.Status, loc)
	}
	return nil
}

func newProjectShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name|id>",
		Short: "Show a single project by name or id",
		Long: `Show one project's name, status, state note, and locations. The argument is
resolved as a name first; if no project has that name and the argument is a
positive integer, it is resolved as a project id.

Examples:
  brag project show bragfile         # by name
  brag project show 3                # by id (when no project is named "3")
  brag project show bragfile --format json`,
		RunE: runProjectShow,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain")
	return cmd
}

func runProjectShow(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("show requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	// LD2: name first, then integer-id fallback.
	project, err := s.GetProjectByName(key)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("get project: %w", err)
		}
		// Name miss: try id iff the key is a positive integer.
		id, convErr := strconv.ParseInt(key, 10, 64)
		if convErr != nil || id <= 0 {
			return UserErrorf("no project named %q", key)
		}
		project, err = s.GetProject(id)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return UserErrorf("no project named or with id %q", key)
			}
			return fmt.Errorf("get project: %w", err)
		}
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectJSON(project)
		if err != nil {
			return fmt.Errorf("render project json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	renderProjectPlain(out, project)
	return nil
}

func renderProjectPlain(out io.Writer, p storage.Project) {
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Status: %s\n", p.Status)
	note := p.StateNote
	if note == "" {
		note = "-"
	}
	fmt.Fprintf(out, "State note: %s\n", note)
	if len(p.Locations) == 0 {
		fmt.Fprintln(out, "Locations: (none)")
		return
	}
	fmt.Fprintln(out, "Locations:")
	for _, l := range p.Locations {
		fmt.Fprintf(out, "  %s\n", l)
	}
}

// resolveProjectByNameOrID resolves key as a project name first, then —
// if no project has that name and key is a positive integer — as a
// project id (mirroring `brag project show`, SPEC-028 LD2). Returns the
// resolved project, or an error wrapping storage.ErrNotFound on a miss.
func resolveProjectByNameOrID(s *storage.Store, key string) (storage.Project, error) {
	project, err := s.GetProjectByName(key)
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return storage.Project{}, err
	}
	id, convErr := strconv.ParseInt(key, 10, 64)
	if convErr != nil || id <= 0 {
		return storage.Project{}, fmt.Errorf("resolve project %q: %w", key, storage.ErrNotFound)
	}
	return s.GetProject(id)
}

func newProjectStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show active projects by recency with brag counts",
		Long: `Show every non-archived project, most-recently-updated first, as a scannable
dashboard: one row per project with its status, total brag count, and state
note. The brag count is the number of brag entries whose project matches the
project name (DEC-017 soft string match). Archived projects are not shown.

Output is plain tab-separated rows (default) or a naked JSON array of status
objects (--format json) per DEC-011. In plain output a long state note is
truncated; the JSON carries it in full.

Examples:
  brag project status                 # name<TAB>status<TAB>count<TAB>state note
  brag project status --format json   # naked JSON array of status objects`,
		RunE: runProjectStatus,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain tab-separated")
	return cmd
}

func runProjectStatus(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	statuses, err := s.ProjectStatuses()
	if err != nil {
		return fmt.Errorf("project statuses: %w", err)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		body, err := export.ToProjectStatusesJSON(statuses)
		if err != nil {
			return fmt.Errorf("render project statuses json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	for _, st := range statuses {
		fmt.Fprintf(out, "%s\t%s\t%d\t%s\n",
			st.Name, st.Status, st.BragCount, truncateStateNote(st.StateNote))
	}
	return nil
}

// truncateStateNote shortens a state note for the plain dashboard so a
// long note doesn't blow out the row. Rune-based so a multibyte note is
// never split mid-character. JSON output is never truncated.
func truncateStateNote(note string) string {
	const max = 50
	r := []rune(note)
	if len(r) <= max {
		return note
	}
	return string(r[:max]) + "…"
}

func newProjectEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name|id>",
		Short: "Edit a project's name, status, or state note",
		Long: `Edit a project's scalar fields. The project is resolved by name first, then
by id. Pass at least one of --name, --status, or --state-note; unspecified
fields are left unchanged.

Renaming a project does NOT rewrite the project string on existing brag entries
— they keep what they were captured with (DEC-017). Editing locations is a
separate command (a later STAGE-007 spec); this edits scalar fields only.

Examples:
  brag project edit bragfile --status paused
  brag project edit bragfile --state-note "shipped tags; next: cut v0.2.0"
  brag project edit bragfile --name brag-cli`,
		RunE: runProjectEdit,
	}
	cmd.Flags().String("name", "", "new project name (rename; must be unique)")
	cmd.Flags().String("status", "", "new status (one of: active, paused, done, archived)")
	cmd.Flags().String("state-note", "", "new state/next-action note")
	return cmd
}

func runProjectEdit(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("edit requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	nameChanged := cmd.Flags().Changed("name")
	statusChanged := cmd.Flags().Changed("status")
	noteChanged := cmd.Flags().Changed("state-note")
	if !nameChanged && !statusChanged && !noteChanged {
		return UserErrorf("edit requires at least one of --name, --status, --state-note")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	current, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	next := current
	if nameChanged {
		name, _ := cmd.Flags().GetString("name")
		name = strings.TrimSpace(name)
		if name == "" {
			return UserErrorf("--name must not be empty")
		}
		next.Name = name
	}
	if statusChanged {
		status, _ := cmd.Flags().GetString("status")
		next.Status = strings.TrimSpace(status)
	}
	if noteChanged {
		note, _ := cmd.Flags().GetString("state-note")
		next.StateNote = note
	}

	updated, err := s.UpdateProject(current.ID, next)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrProjectExists):
			return UserErrorf("project %q already exists", next.Name)
		case errors.Is(err, storage.ErrInvalidStatus):
			return UserErrorf("invalid status %q (accepted: active, paused, done, archived)", next.Status)
		case errors.Is(err, storage.ErrNotFound):
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("update project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Edited project %q.\n", updated.Name)
	return nil
}

func newProjectArchiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <name|id>",
		Short: "Archive a project (non-destructive status flip; recoverable)",
		Long: `Archive a project by setting its status to "archived". This is a
non-destructive flip: the project, its state note, and its locations are all
preserved, and it can be restored at any time with:

  brag project edit <name|id> --status active

Archive is NOT delete. To permanently remove a project and its locations, use
'brag project delete', which is irreversible.

Examples:
  brag project archive bragfile
  brag project archive 3`,
		RunE: runProjectArchive,
	}
	return cmd
}

func runProjectArchive(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("archive requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	project, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	if err := s.ArchiveProject(project.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("archive project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Archived project %q.\n", project.Name)
	return nil
}

func newProjectDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name|id>",
		Short: "Permanently delete a project (irreversible)",
		Long: `Permanently delete a project and its locations. Prompts for confirmation
unless --yes is passed.

This is IRREVERSIBLE and distinct from 'brag project archive' (a recoverable
status flip). Delete removes the project row and every filesystem location
attached to it. Existing brag entries are NOT touched — an entry keeps the
project string it was captured with (DEC-017), so 'brag list --project <name>'
still finds those entries after the project is deleted.

Examples:
  brag project delete bragfile        # prompts y/N on stdin
  brag project delete bragfile --yes  # skip the prompt
  brag project delete 3 -y            # by id, no prompt`,
		RunE: runProjectDelete,
	}
	cmd.Flags().BoolP("yes", "y", false, "skip confirmation prompt")
	return cmd
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return UserErrorf("delete requires exactly one <name|id> argument")
	}
	key := strings.TrimSpace(args[0])
	if key == "" {
		return UserErrorf("project name or id must not be empty")
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	project, err := resolveProjectByNameOrID(s, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("resolve project: %w", err)
	}

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"Delete project %q and its locations? This cannot be undone. [y/N] ", project.Name)
		reader := bufio.NewReader(cmd.InOrStdin())
		line, _ := reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if line != "y" && line != "Y" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	if err := s.DeleteProject(project.ID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return UserErrorf("no project named or with id %q", key)
		}
		return fmt.Errorf("delete project: %w", err)
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Deleted project %q.\n", project.Name)
	return nil
}

func newProjectHereCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "here",
		Short: "Show which project the current directory belongs to",
		Long: `Resolve the current working directory against registered project locations.
Prints the matching project if the cwd is inside a registered location
(nearest-ancestor match — you may be in any subdirectory, not just the exact
registered root). If no registered project matches, exits 1.

Output is a single tab-separated line (default) or a JSON object (--format json)
with the full project shape including locations (same as 'brag project show').

Examples:
  brag project here                 # name<TAB>status<TAB>state-note
  brag project here --format json   # single project JSON object`,
		RunE: runProjectHere,
	}
	cmd.Flags().String("format", "", "output format (one of: json); default is plain")
	return cmd
}

func runProjectHere(cmd *cobra.Command, _ []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "" && format != "json" {
		return UserErrorf("unknown --format value %q (accepted: json)", format)
	}

	cwd, err := getCwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	dbFlag := getFlagString(cmd, "db")
	dbPath, err := config.ResolveDBPath(dbFlag)
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}
	s, err := storage.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	p, err := s.ProjectForPath(cwd)
	if err != nil {
		return fmt.Errorf("resolve project: %w", err)
	}
	if p == nil {
		// Write message explicitly so it appears in cmd.ErrOrStderr() (tests
		// capture this). Return a bare ErrUser so main.go routes to exit 1
		// without double-printing the message.
		fmt.Fprintln(cmd.ErrOrStderr(), "not inside any registered project")
		return fmt.Errorf("%w", ErrUser)
	}

	out := cmd.OutOrStdout()
	if format == "json" {
		full, err := s.GetProject(p.ID)
		if err != nil {
			return fmt.Errorf("get project: %w", err)
		}
		body, err := export.ToProjectJSON(full)
		if err != nil {
			return fmt.Errorf("render project json: %w", err)
		}
		fmt.Fprintln(out, string(body))
		return nil
	}
	// Plain one-liner: name<TAB>status<TAB>state_note (LD2).
	note := p.StateNote
	if note == "" {
		note = "-"
	}
	fmt.Fprintf(out, "%s\t%s\t%s\n", p.Name, p.Status, note)
	return nil
}
