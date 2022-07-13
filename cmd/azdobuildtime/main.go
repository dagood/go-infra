// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/microsoft/go-infra/azdo"
)

func main() {
	if err := run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run() error {
	id := flag.Int(
		"id", 0,
		"The AzDO build ID (not build number) to query. If not passed, interactively\n"+
			"prompts for a build URL to parse to automatically find id, org, and proj.")

	azdoFlags := azdo.BindClientFlags()

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage:\n")
		flag.PrintDefaults()
		fmt.Fprintln(
			flag.CommandLine.Output(),
			"Example usage:\n"+
				"\n"+
				"  azdobuildtime -azdopat $azpat\n"+
				"\n"+
				"  [Paste an AzDO build URL like https://dev.azure.com/dnceng/internal/_build/results?buildId=1876784&view=results and press ENTER]\n"+
				"\n"+
				"It is recommended to use 'read' in Unix shells and 'Read-Host' in PowerShell to\n"+
				"assign '$azpat' rather than passing directly to the command to avoid the PAT\n"+
				"appearing in history.")
	}
	flag.Parse()

	if *id == 0 {
		var scanURL string
		fmt.Print("Build URL: ")
		_, err := fmt.Scan(&scanURL)
		if err != nil {
			return err
		}

		parsedURL, err := url.Parse(scanURL)
		if err != nil {
			return err
		}

		pathParts := strings.Split(parsedURL.Path, "/")
		if len(pathParts) < 3 {
			return fmt.Errorf("pathParts has fewer than 3 parts separated by '/': %v", pathParts)
		}

		org := parsedURL.Scheme + "://" + parsedURL.Host + "/" + pathParts[1] + "/"
		proj := pathParts[2]
		urlID, err := strconv.Atoi(parsedURL.Query().Get("buildId"))
		if err != nil {
			return fmt.Errorf("failed to parse buildId query parameter as int: %w", err)
		}

		azdoFlags.Org = &org
		azdoFlags.Proj = &proj
		id = &urlID
	}

	fmt.Printf("Finding org %q proj %q id %v\n", *azdoFlags.Org, *azdoFlags.Proj, *id)

	ctx := context.Background()
	c, err := build.NewClient(ctx, azdoFlags.NewConnection())
	if err != nil {
		return err
	}
	b, err := c.GetBuild(ctx, build.GetBuildArgs{
		Project:         azdoFlags.Proj,
		BuildId:         id,
		PropertyFilters: nil,
	})
	if err != nil {
		return err
	}

	report(b)

	if url, ok := azdo.GetBuildWebURL(b); ok {
		fmt.Println()
		fmt.Println(url)
	}

	return nil
}

func report(b *build.Build) {
	fmt.Println()

	if b.QueueTime == nil {
		fmt.Println("Build has not been queued.")
		return
	}

	fmt.Printf("queue time: %v\n", b.QueueTime)

	if b.StartTime == nil {
		fmt.Println("Build has not started.")
	}

	fmt.Printf("start time: %v\n", b.StartTime)
	fmt.Println()
	fmt.Printf("  queue -> start (agent acquisiton): %v\n", timesub(b.StartTime, b.QueueTime))

	if b.FinishTime == nil {
		fmt.Println("Build has not finished yet.")
		return
	}

	fmt.Println()
	fmt.Printf("end time: %v\n", b.FinishTime)
	fmt.Println()
	fmt.Printf("  start -> finish: %v\n", timesub(b.FinishTime, b.StartTime))
	fmt.Printf("  queue -> finish: %v\n", timesub(b.FinishTime, b.QueueTime))
}

func timesub(a, b *azuredevops.Time) time.Duration {
	return a.Time.Sub(b.Time).Round(time.Second)
}
