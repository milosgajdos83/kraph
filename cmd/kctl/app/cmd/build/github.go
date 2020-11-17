package build

import (
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/milosgajdos/kraph"
	"github.com/milosgajdos/kraph/pkg/api"
	"github.com/milosgajdos/kraph/pkg/api/gh/star"
	"github.com/milosgajdos/kraph/pkg/store"
	"github.com/milosgajdos/kraph/pkg/store/memory"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

var (
	ghToken  string
	ghUser   string
	ghPaging int
)

// K8s returns K8s subcommand for build command
func GH() *cli.Command {
	return &cli.Command{
		Name:     "kubernetes",
		Aliases:  []string{"k8s"},
		Category: "build",
		Usage:    "kubernetes graph",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "store",
				Aliases:     []string{"s"},
				Value:       "memory",
				Usage:       "graph store",
				Destination: &graphStore,
			},
			&cli.StringFlag{
				Name:        "store-id",
				Aliases:     []string{"id"},
				Value:       "kctl",
				Usage:       "store ID",
				Destination: &graphStoreID,
			},
			&cli.StringFlag{
				Name:        "store-url",
				Aliases:     []string{"u"},
				Value:       "",
				Usage:       "URL of the store",
				EnvVars:     []string{"STORE_URL"},
				Destination: &graphStoreURL,
			},
			&cli.StringFlag{
				Name:        "graph",
				Aliases:     []string{"g"},
				Value:       "owner",
				Usage:       "type of graph",
				Destination: &graphType,
			},
			&cli.StringFlag{
				Name:        "format",
				Aliases:     []string{"f"},
				Value:       "dot",
				Usage:       "print graph in a given format",
				Destination: &graphFormat,
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t"},
				Value:       "",
				Usage:       "GitHub API token",
				EnvVars:     []string{"GITHUB_TOKEN"},
				Destination: &ghToken,
			},
			&cli.StringFlag{
				Name:        "user",
				Aliases:     []string{"u"},
				Value:       "",
				Usage:       "GitHub User",
				Destination: &ghUser,
			},
			&cli.IntFlag{
				Name:        "paging",
				Aliases:     []string{"p"},
				Value:       10,
				Usage:       "GitHub API response paging",
				Destination: &ghPaging,
			},
		},
		Action: func(c *cli.Context) error {
			return runGH(c)
		},
	}
}

func runGH(ctx *cli.Context) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(ctx.Context, ts)

	ghClient := github.NewClient(tc)

	var gstore store.Store
	var err error

	switch graphStore {
	case "memory":
		gstore, err = memory.NewStore(graphStoreID, store.Options{})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported store: %s", graphStore)
	}

	k, err := kraph.New(gstore)
	if err != nil {
		return fmt.Errorf("failed to create kraph: %w", err)
	}

	// TODO: figure this out
	var filters []kraph.Filter

	var client api.Client

	switch graphType {
	case "star":
		client = star.NewClient(ctx.Context, ghClient, star.Paging(ghPaging), star.User(ghUser))
	default:
		return fmt.Errorf("unsupported graph type: %s", graphType)
	}

	if err = k.Build(client, filters...); err != nil {
		return fmt.Errorf("failed to build kraph: %w", err)
	}

	// only print the graph if it's an in-memory graph
	if graphStore == "memory" {
		graphOut, err := graphToOut(k.Store().Graph(), graphFormat)
		if err != nil {
			return err
		}

		fmt.Println(graphOut)
	}

	return nil
}
