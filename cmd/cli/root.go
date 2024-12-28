package cli

import (
	"context"
	"os"

	"github.com/pjbgf/benchr/internal/benchr"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v3"
)

var (
	version  = "(dev)"
	versions []string
	path     string
	target   string
	allocs   string
	ns       string
)

func RootCommand() *cli.Command {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "path",
				Destination: &path,
			},
			&cli.StringFlag{
				Name:        "target",
				Destination: &target,
			},
			&cli.StringSliceFlag{
				Name:        "versions",
				Destination: &versions,
			},
			&cli.StringFlag{
				Name:        "allocs",
				Destination: &allocs,
			},
			&cli.StringFlag{
				Name:        "ns",
				Destination: &ns,
			},
		},
		Action: func(_ context.Context, _ *cli.Command) error {
			if path == "" {
				return errors.New("path cannot be empty")
			}
			if target == "" {
				return errors.New("target cannot be empty")
			}
			if len(versions) == 0 {
				return errors.New("At least one version must be provided")
			}
			b := benchr.New(path, target, versions)

			var opts []benchr.Option
			if allocs != "" {
				af, err := os.Create(allocs)
				if err != nil {
					return errors.Wrap(err, "failed to create file for allocs chart")
				}
				defer af.Close()

				opts = append(opts, benchr.WithAllocsChart(af))
			}

			if ns != "" {
				af, err := os.Create(ns)
				if err != nil {
					return errors.Wrap(err, "failed to create file for ns chart")
				}
				defer af.Close()

				opts = append(opts, benchr.WithNsChart(af))
			}

			return b.Run(opts...)
		},
	}

	cmd.Version = version
	cmd.Usage = "Run Go benchmarks across several version of a dependency"
	cmd.Suggest = true
	cmd.EnableShellCompletion = true

	return cmd
}
