// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/google/go-github/github"
	"github.com/microsoft/go-infra/azdo"
	"github.com/microsoft/go-infra/githubutil"
	"github.com/microsoft/go-infra/subcmd"
)

func init() {
	subcommands = append(subcommands, subcmd.Option{
		Name:        "update-build-list",
		Summary:     "Update the list of builds in the specified GitHub release issue",
		Description: "\n\n" + azdo.AzDOBuildDetectionDoc,
		Handle:      handleUpdateBuildList,
	})
}

func handleUpdateBuildList(p subcmd.ParseFunc) error {
	repo := githubutil.BindRepoFlag()
	pat := githubutil.BindPATFlag()
	issue := flag.Int("i", 0, "[Required] The issue number to add the comment to.")
	status := flag.String("status", "", "[Required] The current Agent.JobStatus value.")

	if err := p(); err != nil {
		return err
	}

	if *issue == 0 {
		return errors.New("no issue specified")
	}
	if *status == "" {
		return errors.New("no status specified")
	}
	owner, name, err := githubutil.ParseRepoFlag(repo)
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := githubutil.NewClient(ctx, *pat)
	if err != nil {
		return err
	}

	if url := azdo.GetEnvBuildURL(); url != "" {
		*message = "[" + azdo.GetEnvBuildID() + "](" + url + "): " + *message

		if *instructionsLink {
			*message = *message + "\n" +
				"[Click here to see " + azdo.GetEnvBuildID() + " retry instructions.](" +
				azdo.GetEnvBuildURL() + "&view=ms.vss-build-web.run-extensions-tab" +
				")"
		}
	}

	var body string
	if err = githubutil.Retry(func() error {
		issue, _, err := client.Issues.Get(ctx, owner, name, issue)
		if err != nil {
			return err
		}
		body = issue.GetBody()
		return nil
	}); err != nil {
		return err
	}

	return githubutil.Retry(func() error {
		c, _, err := client.Issues.CreateComment(
			ctx, owner, name, *issue, &github.IssueComment{Body: body})
		if err != nil {
			return err
		}
		log.Printf("Comment: %v\n", *c.HTMLURL)
		return nil
	})
}
