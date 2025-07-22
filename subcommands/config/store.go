package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/storage"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/utils"

	"gopkg.in/yaml.v3"
)

type ConfigStoreCmd struct {
	subcommands.SubcommandBase

	args []string
}

func (cmd *ConfigStoreCmd) Parse(ctx *appcontext.AppContext, args []string) error {
	flags := flag.NewFlagSet("store", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s\n", flags.Name())
		flags.PrintDefaults()
	}

	flags.Parse(args)
	cmd.args = flags.Args()

	return nil
}

func (cmd *ConfigStoreCmd) Execute(ctx *appcontext.AppContext, repo *repository.Repository) (int, error) {

	err := cmd_store_config(ctx, cmd.args)
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func cmd_store_config(ctx *appcontext.AppContext, args []string) error {
	usage := "usage: plakar store [add|check|ls|ping|rm|set|unset]"
	cmd := "ls"
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	switch cmd {
	case "add":
		usage := "usage: plakar store add <name> <location> [<key>=<value>, ...]"
		if len(args) < 2 {
			return fmt.Errorf(usage)
		}
		name, location := args[0], normalizeLocation(args[1])
		if ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q already exists", name)
		}
		ctx.Config.Repositories[name] = make(map[string]string)
		ctx.Config.Repositories[name]["location"] = location
		for _, kv := range args[2:] {
			key, val, found := strings.Cut(kv, "=")
			if !found {
				return fmt.Errorf(usage)
			}
			if key == "" {
				return fmt.Errorf(usage)
			}
			ctx.Config.Repositories[name][key] = val
		}
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)

	case "check":
		usage := "usage: plakar store check <name>"
		if len(args) != 1 {
			return fmt.Errorf(usage)
		}
		name := args[0]
		if !ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q does not exists", name)
		}
		store, err := storage.New(ctx.GetInner(), ctx.Config.Repositories[name])
		if err != nil {
			return err
		}
		err = store.Close()
		if err != nil {
			ctx.GetLogger().Warn("error when closing store: %v", err)
		}
		return nil

	case "default":
		usage := "usage: plakar store default <name>"
		if len(args) != 1 {
			return fmt.Errorf(usage)
		}
		name := args[0]
		if !ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q does not exists", name)
		}
		ctx.Config.DefaultRepository = name
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)

	case "ls":
		usage := "usage: plakar store ls"
		if len(args) != 0 {
			return fmt.Errorf(usage)
		}
		return yaml.NewEncoder(ctx.Stdout).Encode(ctx.Config.Repositories)

	case "ping":
		return fmt.Errorf("not implemented")

	case "rm":
		usage := "usage: plakar store rm <name>"
		if len(args) != 1 {
			return fmt.Errorf(usage)
		}
		name := args[0]
		if !ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q does not exist", name)
		}
		delete(ctx.Config.Repositories, name)
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)

	case "set":
		usage := "usage: plakar store set <name> [<key>=<value>, ...]"
		if len(args) == 0 {
			return fmt.Errorf(usage)
		}
		name := args[0]
		if !ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q does not exists", name)
		}
		for _, kv := range args[1:] {
			key, val, found := strings.Cut(kv, "=")
			if !found {
				return fmt.Errorf(usage)
			}
			if key == "" {
				return fmt.Errorf(usage)
			}
			ctx.Config.Repositories[name][key] = val
		}
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)

	case "unset":
		usage := "usage: plakar store unset <name> [<key>, ...]"
		if len(args) == 0 {
			return fmt.Errorf(usage)
		}
		name := args[0]
		if !ctx.Config.HasRepository(name) {
			return fmt.Errorf("store %q does not exists", name)
		}
		for _, key := range args[1:] {
			if key == "location" {
				return fmt.Errorf("cannot unset location")
			}
			delete(ctx.Config.Repositories[name], key)
		}
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)
	case "import":
		usage := "usage: plakar store import --format <name> <path_to_location>"
		if len(args) < 3 {
			return fmt.Errorf(usage)
		}
		names := args[2:]
		format := args[0]
		switch format {
		case "-ini":
			iniMap, err := utils.LoadIni(args[1])
			if err != nil {
				return fmt.Errorf("failed to load ini config: %w", err)
			}
			for _, name := range names {
				if ctx.Config.HasRepository(name) {
					fmt.Printf("store %q already exists, skipping\n", name)
					continue
				}
				err := utils.ImportConfigFromIni(ctx, name, iniMap, "store")
				if err != nil {
					fmt.Printf("failed to import store from ini: %w", err)
					continue
				}
			}
		default:
			return fmt.Errorf("unknown format %q. Known formats are: -ini", format)
		}
		return utils.SaveConfig(ctx.ConfigDir, ctx.Config)
	default:
		return fmt.Errorf(usage)
	}
}
