// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/microsoft/go-infra/azdo"
	"github.com/microsoft/go-infra/buildreport"
	"github.com/microsoft/go-infra/githubutil"
	"github.com/microsoft/go-infra/subcmd"
)

func init() {
	subcommands = append(subcommands, subcmd.Option{
		Name:        "report",
		Summary:     "Report release build's status to a GitHub issue",
		Description: "\n\n" + azdo.AzDOBuildDetectionDoc,
		Handle:      handleReport,
	})
}

func handleReport(p subcmd.ParseFunc) error {
	repo := githubutil.BindRepoFlag()
	pat := githubutil.BindPATFlag()
	issue := flag.Int("i", 0, "[Required] The issue number to add the comment to.")

	fromEnv := flag.Bool(
		"azdoenv", false,
		"Gather information from Azure Pipelines predefined env variables.\n"+
			"If more specific info flags are used they override this env info.")

	status := flag.String("build-status", "", "The current Agent.JobStatus value.")
	buildID := flag.String("build-id", "", "The build ID to report.")
	buildPipeline := flag.String("build-pipeline", "", "The name of the build pipeline.")
	start := flag.Bool("build-start", false, "Assign the current time as the start time of the reported build.")

	version := flag.String(
		"version", "",
		"A full microsoft/go version number (major.minor.patch-revision[-suffix]).\n"+
			"This is used to categorize the list of builds in a release issue.")

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

	s := buildreport.State{
		Version:    *version,
		LastUpdate: time.Now().UTC(),
	}
	buildStatus := ""
	if *fromEnv {
		s.BuildID = azdo.GetEnvBuildID()
		s.BuildPipeline = azdo.GetEnvDefinitionName()
		s.BuildURL = azdo.GetEnvBuildURL()
		buildStatus = azdo.GetEnvAgentJobStatus()
	}
	if *buildID != "" {
		s.BuildID = *buildID
	}
	if *buildPipeline != "" {
		s.BuildPipeline = *buildPipeline
	}
	if *status != "" {
		buildStatus = *status
	}
	if *start {
		s.StartTime = s.LastUpdate
	}

	switch buildStatus {
	// Handle possible AzDO env values.
	case "Succeeded", "SucceededWithIssues":
		s.BuildSymbol = buildreport.SuccessSymbol
	case "Canceled", "Failed":
		s.BuildSymbol = buildreport.FailedSymbol
	// Handle values that are passed in manually, not provided by Agent.JobStatus.
	case "InProgress":
		s.BuildSymbol = buildreport.InProgressSymbol
	case "NotStarted":
		s.BuildSymbol = buildreport.NotStartedSymbol
	}

	log.Printf("Reporting %#v\n", s)

	ctx := context.Background()

	if err := buildreport.UpdateIssue(ctx, owner, name, *pat, *issue, s); err != nil {
		return err
	}

	client, err := githubutil.NewClient(ctx, *pat)
	if err != nil {
		return err
	}

	if err := buildreport.UpdateIssue(ctx, owner, name, *pat, *issue, s); err != nil {
		return err
	}

	if s.BuildSymbol == buildreport.FailedSymbol {
		if err := githubutil.Retry(func() error {
			body, err := buildreport.SingleStateBody(&s)
			if err != nil {
				return err
			}

			updateMessage := "Build failure: " + body

			c, _, err := client.Issues.CreateComment(
				ctx, owner, name, *issue, &github.IssueComment{Body: &updateMessage})
			if err != nil {
				return err
			}
			log.Printf("Comment: %v\n", *c.HTMLURL)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
