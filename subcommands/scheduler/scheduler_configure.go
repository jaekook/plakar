package scheduler

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/scheduler"
	"github.com/PlakarKorp/plakar/subcommands"
)

type SchedulerConfigure struct {
	subcommands.SubcommandBase

	socketPath string

	configBytes []byte
}

func (cmd *SchedulerConfigure) Parse(ctx *appcontext.AppContext, args []string) error {
	flags := flag.NewFlagSet("agent tasks configure", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage: %s [OPTIONS]\n", flags.Name())
		fmt.Fprintf(flags.Output(), "\nOPTIONS:\n")
		flags.PrintDefaults()
	}
	flags.Parse(args)

	if flags.NArg() == 0 {
		flags.Usage()
		return fmt.Errorf("no configuration file provided")
	}
	if flags.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}

	configurationLocation := flags.Arg(0)

	var rd io.Reader

	if strings.HasPrefix(configurationLocation, "http://") || strings.HasPrefix(configurationLocation, "https://") {
		resp, err := http.Get(configurationLocation)
		if err != nil {
			return fmt.Errorf("failed to download configuration file from %q: %w", configurationLocation, err)
		}
		defer resp.Body.Close()
		rd = resp.Body
	} else {
		absolutePath, err := filepath.Abs(configurationLocation)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for configuration file: %w", err)
		}
		fp, err := os.Open(absolutePath)
		if err != nil {
			return fmt.Errorf("failed to open configuration file %q: %w", absolutePath, err)
		}
		defer fp.Close()
		rd = fp
	}

	configBytes, err := io.ReadAll(rd)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}
	cmd.configBytes = configBytes
	cmd.socketPath = filepath.Join(ctx.CacheDir, "scheduler.sock")
	return nil
}

func (cmd *SchedulerConfigure) Execute(ctx *appcontext.AppContext, repo *repository.Repository) (int, error) {
	cl, err := scheduler.NewClient(cmd.socketPath, false)
	if err != nil {
		if err == scheduler.ErrWrongVersion {
			return 1, fmt.Errorf("scheduler is running with a different version of plakar: %w", err)
		}
		return 1, fmt.Errorf("failed to connect to scheduler: %w", err)
	}
	defer cl.Close()

	return cl.Configure(cmd.configBytes)
}
