package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/xerrors"

	"github.com/KyleBanks/depth"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func main() {
	// ターミナルに出力する場合のみカラー出力する
	color.NoColor = !isatty.IsTerminal(os.Stdout.Fd())

	if succeeded, err := run(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	} else if succeeded {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func run() (bool, error) {
	config, err := NewConfig()
	if err != nil {
		return false, xerrors.Errorf(": %w", err)
	}

	resolver, err := NewPathResolver(config.AbsPath)
	if err != nil {
		return false, xerrors.Errorf(": %w", err)
	}

	sem := semaphore.NewWeighted(config.Concurrency)
	eg := errgroup.Group{}
	showMutex := sync.Mutex{}
	foundError := false

	for _, c := range config.Constraint {
		// "./hoge/fuga" 形式を "github.com/tomoemon/hoge/fuga" 形式に統一する
		fromOriginalPath := resolver.NormalizeImportPath(c.From)
		allowedPackages := resolver.NormalizeImportPaths(c.Allow)

		fromPaths, err := resolver.ExpandWildCardSuffix(fromOriginalPath)
		if err != nil {
			return false, xerrors.Errorf(": %w", err)
		}
		for _, fromPath := range fromPaths {
			c := c
			fromPath := fromPath
			eg.Go(func() error {
				if err := sem.Acquire(context.Background(), 1); err != nil {
					return xerrors.Errorf(": %w", err)
				}
				defer sem.Release(1)

				result, err := resolve(fromPath, allowedPackages, resolver.ModuleName(), config)
				if xerrors.Is(err, depth.ErrRootPkgNotResolved) {
					if len(fromPaths) > 1 {
						return nil
					}
					return xerrors.Errorf("from \"%s\" (\"%s\") cannot be resolved: %w", c.From, fromOriginalPath, err)
				} else if err != nil {
					return xerrors.Errorf(": %w", err)
				}

				showMutex.Lock()
				fmt.Printf("# %s\n", fromPath)
				for _, r := range result {
					if r.err != nil {
						printResult(false, r.err.Error())
						foundError = true
					} else {
						printResult(true, r.name)
					}
				}
				fmt.Printf("\n")
				showMutex.Unlock()

				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		return false, xerrors.Errorf(": %w", err)
	}
	if foundError {
		return false, nil
	}
	return true, nil
}

func printResult(isOk bool, message string) {
	if isOk {
		c := color.New(color.FgGreen)
		c.Printf("[OK] %s\n", message)
	} else {
		c := color.New(color.FgRed)
		c.Printf("[NG] %s\n", message)
	}
}
