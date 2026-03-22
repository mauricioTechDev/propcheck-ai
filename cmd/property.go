package cmd

import (
	"fmt"
	"strconv"

	"github.com/mauricioTechDev/propcheck-ai/internal/session"
	"github.com/mauricioTechDev/propcheck-ai/internal/types"
	"github.com/spf13/cobra"
)

var propertyCmd = &cobra.Command{
	Use:   "property",
	Short: "Manage PBT properties",
	Long:  "Add, list, or complete properties that define what to test.",
	Example: `  propcheck-ai property add "sort is idempotent"
  propcheck-ai property list
  propcheck-ai property pick 1
  propcheck-ai property done 1`,
}

var propertyCategoryFlag string

var propertyAddCmd = &cobra.Command{
	Use:   "add \"description\"",
	Short: "Add a new property to test",
	Long:  "Add one or more properties to the current PBT session. Each argument is a separate property description.",
	Example: `  propcheck-ai property add "sort is idempotent: sort(sort(xs)) == sort(xs)"
  propcheck-ai property add --category roundtrip "parse(format(x)) == x"
  propcheck-ai property add "reverse is its own inverse" "append preserves length"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if propertyCategoryFlag != "" {
			valid := false
			for _, c := range types.ValidCategories() {
				if propertyCategoryFlag == c {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid category %q. Valid: invariant, roundtrip, equivalence, metamorphic", propertyCategoryFlag)
			}
		}

		for _, desc := range args {
			id := s.AddProperty(desc, propertyCategoryFlag)
			if propertyCategoryFlag != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Property [%d] added (%s): %s\n", id, propertyCategoryFlag, desc)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Property [%d] added: %s\n", id, desc)
			}
		}

		s.AddEvent("property_add", func(e *types.Event) {
			e.PropCount = len(args)
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}
		return nil
	},
}

var propertyListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all properties",
	Long:    "Display all properties in the current session with their status.",
	Example: `  propcheck-ai property list`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if len(s.Properties) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No properties defined. Add properties with 'propcheck-ai property add \"desc\"'")
			return nil
		}

		for _, p := range s.Properties {
			status := "active"
			if p.Status == types.PropertyStatusCompleted {
				status = "done"
			}
			cat := ""
			if p.Category != "" {
				cat = fmt.Sprintf(" [%s]", p.Category)
			}
			isCurrent := s.CurrentPropertyID != nil && p.ID == *s.CurrentPropertyID
			if isCurrent {
				fmt.Fprintf(cmd.OutOrStdout(), "→ [%d] (%s)%s %s (current)\n", p.ID, status, cat, p.Description)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "  [%d] (%s)%s %s\n", p.ID, status, cat, p.Description)
			}
		}
		return nil
	},
}

var propertyDoneAll bool

var propertyDoneCmd = &cobra.Command{
	Use:   "done <id> [id...]",
	Short: "Mark a property as completed",
	Long:  "Mark one or more properties as completed by their ID. Use --all to mark every active property as done.",
	Example: `  propcheck-ai property done 1
  propcheck-ai property done 1 2 3
  propcheck-ai property done --all`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		if propertyDoneAll && len(args) > 0 {
			return fmt.Errorf("cannot use --all with specific property IDs")
		}
		if !propertyDoneAll && len(args) == 0 {
			return fmt.Errorf("provide at least one property ID, or use --all")
		}

		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if propertyDoneAll {
			count := s.CompleteAllProperties()
			if count == 0 {
				return fmt.Errorf("no active properties to mark as done")
			}
			s.AddEvent("property_done", func(e *types.Event) {
				e.PropCount = count
			})
			if err := session.Save(dir, s); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Marked %d property(ies) as done\n", count)
			return nil
		}

		for _, arg := range args {
			id, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("property ID must be a number, got %q", arg)
			}
			if err := s.CompleteProperty(id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Property [%d] marked as done\n", id)
		}

		s.AddEvent("property_done", func(e *types.Event) {
			e.PropCount = len(args)
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}
		return nil
	},
}

var propertyPickCmd = &cobra.Command{
	Use:   "pick <id>",
	Short: "Pick a property to work on in this iteration",
	Long:  "Select an active property to focus on for the current PBT iteration.",
	Example: `  propcheck-ai property pick 1
  propcheck-ai property pick 3`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getWorkDir()
		s, err := session.LoadOrFail(dir)
		if err != nil {
			return err
		}

		if s.Phase != types.PhaseProperty {
			return fmt.Errorf("can only pick a property during the PROPERTY phase (current phase: %s)", s.Phase)
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("property ID must be a number, got %q", args[0])
		}

		if err := s.SetCurrentProperty(id); err != nil {
			return err
		}

		s.AddEvent("property_picked", func(e *types.Event) {
			e.PropertyID = id
		})

		if err := session.Save(dir, s); err != nil {
			return err
		}

		p := s.CurrentProperty()
		remaining := s.RemainingProperties()
		fmt.Fprintf(cmd.OutOrStdout(), "Picked property [%d]: %s\n", p.ID, p.Description)
		fmt.Fprintf(cmd.OutOrStdout(), "%d property(ies) remaining after this one\n", len(remaining))
		return nil
	},
}

func init() {
	propertyAddCmd.Flags().StringVar(&propertyCategoryFlag, "category", "", "property category: invariant, roundtrip, equivalence, metamorphic")
	propertyDoneCmd.Flags().BoolVar(&propertyDoneAll, "all", false, "mark all active properties as done")
	propertyCmd.AddCommand(propertyAddCmd)
	propertyCmd.AddCommand(propertyListCmd)
	propertyCmd.AddCommand(propertyDoneCmd)
	propertyCmd.AddCommand(propertyPickCmd)
	rootCmd.AddCommand(propertyCmd)
}
