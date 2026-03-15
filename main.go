package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/shv-ng/frec/internal/cmd"
	"github.com/urfave/cli/v3"
)

var Version string = "dev"

func main() {
	app := &cli.Command{
		Name:                  "frec",
		Usage:                 "track items by frequency and recency",
		Commands:              []*cli.Command{},
		EnableShellCompletion: true,
		Version:               Version,
	}

	versionCmd := &cli.Command{
		Name:  "version",
		Usage: "see version",
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Printf("%s version %s\n", app.Name, app.Version)
			return nil
		},
	}
	app.Commands = append(app.Commands, versionCmd)

	addCmd := &cli.Command{
		Name:      "add",
		Usage:     "add item to namespace",
		ArgsUsage: "<namespace> <item>",
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			item := c.Args().Get(1)
			if ns == "" || item == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.AddItem(ns, item)
		},
	}
	app.Commands = append(app.Commands, addCmd)

	listCmd := &cli.Command{
		Name:      "list",
		Usage:     "list items from namespace",
		ArgsUsage: "<namespace> <options>",
		Aliases:   []string{"ls"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "${name}\t${star}\tscore:${score}\tvisits:${count}\tlast:${lastseen}d ago\n",
				Usage: "format rows",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: -1,
				Usage: "top N items",
				Validator: func(i int) error {
					if i >= 0 {
						return nil
					}
					return fmt.Errorf("limit can't be negative, found: %d", i)
				},
			},
			&cli.IntFlag{
				Name:  "since",
				Value: -1,
				Usage: "visited in last N days",
				Validator: func(i int) error {
					if i >= 0 {
						return nil
					}
					return fmt.Errorf("since days can't be negative, found: %d", i)
				},
			},
			&cli.BoolFlag{
				Name:  "starred",
				Value: false,
				Usage: "only starred items",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			if ns == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.ListItems(ns, c.String("format"), c.Int("limit"), c.Int("since"), c.Bool("starred"))
		},
	}
	app.Commands = append(app.Commands, listCmd)

	removeCmd := &cli.Command{
		Name:      "remove",
		Usage:     "remove item to namespace",
		ArgsUsage: "<namespace> <item>",
		Aliases:   []string{"rm"},
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			item := c.Args().Get(1)
			if ns == "" || item == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.RemoveItem(ns, item)
		},
	}
	app.Commands = append(app.Commands, removeCmd)

	starCmd := &cli.Command{
		Name:      "star",
		Usage:     "star a item to namespace",
		ArgsUsage: "<namespace> <item>",
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			item := c.Args().Get(1)
			if ns == "" || item == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.StarItem(ns, item, true)
		},
	}
	app.Commands = append(app.Commands, starCmd)

	unstarCmd := &cli.Command{
		Name:      "unstar",
		Usage:     "unstar a item to namespace",
		ArgsUsage: "<namespace> <item>",
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			item := c.Args().Get(1)
			if ns == "" || item == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.StarItem(ns, item, false)
		},
	}
	app.Commands = append(app.Commands, unstarCmd)

	nsCmd := &cli.Command{
		Name:     "namespace",
		Usage:    "list/remove namespace",
		Aliases:  []string{"ns"},
		Commands: []*cli.Command{},
	}

	nslistCmd := &cli.Command{
		Name:    "list",
		Usage:   "list namespace",
		Aliases: []string{"ls"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "format",
				Value: "namespace: ${ns}\n",
				Usage: "format rows",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return cmd.ListNs(c.String("format"))
		},
	}
	nsCmd.Commands = append(nsCmd.Commands, nslistCmd)

	nsRemoveCmd := &cli.Command{
		Name:      "remove",
		Usage:     "remove namespace",
		Aliases:   []string{"rm"},
		ArgsUsage: "<namespace>",
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			if ns == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			reader := bufio.NewReaderSize(os.Stdin, 1)
			fmt.Printf("Are you sure want to remove %s? [y/N] ", ns)
			ch, err := reader.ReadByte()
			if err != nil {
				return err
			}
			if ch != 'y' && ch != 'Y' {
				return nil
			}
			return cmd.RemoveNs(ns)
		},
	}
	nsCmd.Commands = append(nsCmd.Commands, nsRemoveCmd)

	app.Commands = append(app.Commands, nsCmd)

	syncCmd := &cli.Command{
		Name:      "sync",
		Usage:     "sync items in namespace from stdin",
		ArgsUsage: "<namespace>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "null",
				Usage: "seperate by null character (\\0)",
				Value: false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			ns := c.Args().Get(0)
			if ns == "" {
				cli.ShowSubcommandHelpAndExit(c, 1)
			}
			return cmd.SyncItems(ns, c.Bool("null"))
		},
	}
	app.Commands = append(app.Commands, syncCmd)

	app.Run(context.Background(), os.Args)
}
