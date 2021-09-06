package rpflag

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	kConflicts  = "conflicts"
	kDepends    = "depends"
	kOneAtLeast = "oneatleast"
	kMandatory  = "mandatory"
)

func Conflicts(cmd *cobra.Command, flags ...string) {
	for i := range flags {
		if f := cmd.Flag(flags[i]); f != nil {
			if f.Annotations == nil {
				f.Annotations = make(map[string][]string)
			}

			xs := make([]string, len(flags)-1)
			copy(xs, flags[:i])
			copy(xs[i:], flags[i+1:])

			f.Annotations[kConflicts] = append(f.Annotations[kConflicts], xs...)
		}
	}
}

func Depends(cmd *cobra.Command, flags ...string) {
	for i := range flags {
		if f := cmd.Flag(flags[i]); f != nil {
			if f.Annotations == nil {
				f.Annotations = make(map[string][]string)
			}

			xs := make([]string, len(flags)-1)
			copy(xs, flags[:i])
			copy(xs[i:], flags[i+1:])

			f.Annotations[kDepends] = append(f.Annotations[kDepends], xs...)
		}
	}
}

func OneAtLeast(cmd *cobra.Command) {
	cmd.Annotations = make(map[string]string)
	cmd.Annotations[kOneAtLeast] = ""
}

func Mandatory(cmd *cobra.Command, flags ...string) {
	for i := range flags {
		if f := cmd.Flag(flags[i]); f != nil {
			if f.Annotations == nil {
				f.Annotations = make(map[string][]string)
			}

			f.Annotations[kMandatory] = nil
		}
	}
}

func Resolve(cmd *cobra.Command) error {
	if err := resolveMandatory(cmd); err != nil {
		return err
	}

	if err := resolveOneAtLeast(cmd); err != nil {
		return err
	}

	if err := resolveConflicts(cmd); err != nil {
		return err
	}

	if err := resolveDepends(cmd); err != nil {
		return err
	}

	return nil
}

func resolveConflicts(cmd *cobra.Command) (err error) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		for _, n := range flag.Annotations[kConflicts] {
			if other := cmd.Flag(n); other != nil && other.Changed {
				err = fmt.Errorf(
					"\"-%s, --%s\" conflicts with \"-%s, --%s\"",
					flag.Shorthand, flag.Name,
					other.Shorthand, other.Name)
				break
			}
		}
	})

	return
}

func resolveDepends(cmd *cobra.Command) (err error) {
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		for _, n := range flag.Annotations[kDepends] {
			if other := cmd.Flag(n); other != nil && !other.Changed {
				err = fmt.Errorf(
					"\"-%s, --%s\" depends on \"-%s, --%s\"",
					flag.Shorthand, flag.Name,
					other.Shorthand, other.Name)
				break
			}
		}
	})

	return
}

func resolveOneAtLeast(cmd *cobra.Command) (err error) {
	if _, ok := cmd.Annotations[kOneAtLeast]; ok {
		var flags []string
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			flags = append(flags, fmt.Sprintf("\"-%s, --%s\"", flag.Shorthand, flag.Name))
		})

		err = fmt.Errorf("one or more flags required: %s", strings.Join(flags, ", "))
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			err = nil
		})
	}
	return
}

func resolveMandatory(cmd *cobra.Command) (err error) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if _, ok := flag.Annotations[kMandatory]; ok && !flag.Changed {
			err = fmt.Errorf(
				"mandatory flag \"-%s, --%s\" not specified",
				flag.Shorthand, flag.Name,
			)
			return
		}
	})

	return
}
