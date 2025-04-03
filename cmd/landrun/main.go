package main

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/zouuup/landrun/internal/exec"
	"github.com/zouuup/landrun/internal/log"
	"github.com/zouuup/landrun/internal/sandbox"
)

// Version is the current version of landrun
const Version = "0.1.13"

func main() {
	app := &cli.App{
		Name:    "landrun",
		Usage:   "Run a command in a Landlock sandbox",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Set logging level (error, info, debug)",
				Value:   "error",
				EnvVars: []string{"LANDRUN_LOG_LEVEL"},
			},
			&cli.StringSliceFlag{
				Name:  "ro",
				Usage: "Allow read-only access to this path",
			},
			&cli.StringSliceFlag{
				Name:  "rox",
				Usage: "Allow read-only access with execution to this path",
			},
			&cli.StringSliceFlag{
				Name:  "rw",
				Usage: "Allow read-write access to this path",
			},
			&cli.StringSliceFlag{
				Name:  "rwx",
				Usage: "Allow read-write access with execution to this path",
			},
			&cli.IntSliceFlag{
				Name:   "bind-tcp",
				Usage:  "Allow binding to these TCP ports",
				Hidden: false,
			},
			&cli.IntSliceFlag{
				Name:   "connect-tcp",
				Usage:  "Allow connecting to these TCP ports",
				Hidden: false,
			},
			&cli.BoolFlag{
				Name:  "best-effort",
				Usage: "Use best effort mode (fall back to less restrictive sandbox if necessary)",
				Value: false,
			},
			&cli.StringSliceFlag{
				Name:  "env",
				Usage: "Environment variables to pass to the sandboxed command (KEY=VALUE or just KEY to pass current value)",
				Value: cli.NewStringSlice(),
			},
			&cli.BoolFlag{
				Name:  "unrestricted-filesystem",
				Usage: "Allow unrestricted filesystem access",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "unrestricted-network",
				Usage: "Allow unrestricted network access",
				Value: false,
			},
		},
		Before: func(c *cli.Context) error {
			log.SetLevel(c.String("log-level"))
			return nil
		},
		Action: func(c *cli.Context) error {
			args := c.Args().Slice()
			if len(args) == 0 {
				log.Fatal("Missing command to run")
			}

			// Combine --ro and --rox paths for read-only access
			readOnlyPaths := append([]string{}, c.StringSlice("ro")...)
			readOnlyPaths = append(readOnlyPaths, c.StringSlice("rox")...)

			// Combine --rw and --rwx paths for read-write access
			readWritePaths := append([]string{}, c.StringSlice("rw")...)
			readWritePaths = append(readWritePaths, c.StringSlice("rwx")...)

			// Combine --rox and --rwx paths for executable permissions
			readOnlyExecutablePaths := append([]string{}, c.StringSlice("rox")...)
			readWriteExecutablePaths := append([]string{}, c.StringSlice("rwx")...)

			cfg := sandbox.Config{
				ReadOnlyPaths:            readOnlyPaths,
				ReadWritePaths:           readWritePaths,
				ReadOnlyExecutablePaths:  readOnlyExecutablePaths,
				ReadWriteExecutablePaths: readWriteExecutablePaths,
				BindTCPPorts:             c.IntSlice("bind-tcp"),
				ConnectTCPPorts:          c.IntSlice("connect-tcp"),
				BestEffort:               c.Bool("best-effort"),
				UnrestrictedFilesystem:   c.Bool("unrestricted-filesystem"),
				UnrestrictedNetwork:      c.Bool("unrestricted-network"),
			}

			// Process environment variables
			envVars := processEnvironmentVars(c.StringSlice("env"))

			if err := sandbox.Apply(cfg); err != nil {
				log.Fatal("Failed to apply sandbox: %v", err)
			}

			return exec.Run(args, envVars)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("%v", err)
	}
}

// processEnvironmentVars processes the env flag values
func processEnvironmentVars(envFlags []string) []string {
	result := []string{}

	for _, env := range envFlags {
		// If the flag is just a key (no = sign), get the value from the current environment
		if !strings.Contains(env, "=") {
			if val, exists := os.LookupEnv(env); exists {
				result = append(result, env+"="+val)
			}
		} else {
			// Flag already contains the value (KEY=VALUE format)
			result = append(result, env)
		}
	}

	return result
}
