// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/go-github/github"
	"github.com/microsoft/go-infra/azdo"
	"github.com/microsoft/go-infra/buildreport"
	"github.com/microsoft/go-infra/githubutil"
	"github.com/microsoft/go-infra/subcmd"
)

func init() {
	subcommands = append(subcommands, subcmd.Option{
		Name:    "report",
		Summary: "Report release build's status to a GitHub issue",
		Description: `

By default, this task uses the build environment to report the current build's status to the target
issue. These values can be overridden by passing additional args to report the status of another
build, such as when a new build is queued, or to specify a more exact build status like "InProgress"
when that info isn't clearly in the environment.
`,
		Handle: handleReport,
	})
}

func handleReport(p subcmd.ParseFunc) error {
	repo := githubutil.BindRepoFlag()
	pat := githubutil.BindPATFlag()
	issue := flag.Int("i", 0, "[Required] The issue number to add the comment to.")

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

	owner, name, err := githubutil.ParseRepoFlag(repo)
	if err != nil {
		return err
	}

	s := buildreport.State{
		Version:    *version,
		Name:       azdo.GetEnvDefinitionName(),
		ID:         azdo.GetEnvBuildID(),
		LastUpdate: time.Now().UTC(),
	}
	if *buildPipeline != "" {
		s.Name = *buildPipeline
	}
	if *buildID != "" {
		s.ID = *buildID
	}
	if *start {
		s.StartTime = s.LastUpdate
	}
	// Assume we're always reporting on a build in the same collection/project as the current build.
	// If this is no longer true
	s.URL = azdo.GetBuildURL(azdo.GetEnvCollectionURI(), azdo.GetEnvProject(), s.ID)

	buildStatus := azdo.GetEnvAgentJobStatus()
	if *status != "" {
		buildStatus = *status
	}

	switch buildStatus {
	// Handle possible AzDO env values.
	case "Succeeded", "SucceededWithIssues":
		s.Symbol = buildreport.ReportSymbolSucceeded
	case "Canceled", "Failed":
		s.Symbol = buildreport.ReportSymbolFailed
	// Handle values that are passed in manually, not provided by Agent.JobStatus.
	case "InProgress":
		s.Symbol = buildreport.ReportSymbolInProgress
	case "NotStarted":
		s.Symbol = buildreport.ReportSymbolNotStarted
	}

	log.Printf("Reporting %#v\n", s)

	ctx := context.Background()

	//if err := buildreport.UpdateIssue(ctx, owner, name, *pat, *issue, s); err != nil {
	//	return err
	//}

	client, err := githubutil.NewClient(ctx, *pat)
	if err != nil {
		return err
	}

	log.Println("Before")
	d := func() {
		log.Println("Hi")
		s2 := s
		s2.ID += strconv.Itoa(rand.Int())
		s2.Name += strconv.Itoa(rand.Int())
		if err := buildreport.UpdateIssue(ctx, owner, name, *pat, *issue, s2); err != nil {
			log.Fatal(err)
			//return err
		}
	}
	d()
	//go d()
	//go d()
	//go d()
	//time.Sleep(time.Minute * 5)

	if s.Symbol == buildreport.ReportSymbolFailed {
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
