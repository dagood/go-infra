// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/microsoft/go-infra/subcmd"
)

func init() {
	subcommands = append(subcommands, subcmd.Option{
		Name:    "list",
		Summary: "List known MinGW sources and versions.",
		Handle:  list,
	})
}

func list(p subcmd.ParseFunc) error {
	initFilterFlags()
	format := flag.String("format", "", "a custom Go template used to produce each line of output")
	unique := flag.Bool("unique", false, "only print unique lines")
	if err := p(); err != nil {
		return err
	}

	var tmpl *template.Template
	if *format != "" {
		log.Printf("Using custom format %#q", *format)
		var err error
		tmpl, err = template.New("").Parse(*format)
		if err != nil {
			return err
		}
	}

	existingBuilds, err := unmarshal()
	if err != nil {
		return err
	}
	builds := filter(existingBuilds)

	var printed map[string]struct{}
	if *unique {
		printed = make(map[string]struct{})
	}
	for _, b := range builds {
		var line strings.Builder
		if tmpl != nil {
			if err := tmpl.Execute(&line, b); err != nil {
				return err
			}
		} else {
			fmt.Fprintf(&line, "%#v", b)
		}
		if *unique {
			if _, ok := printed[line.String()]; ok {
				continue
			}
			printed[line.String()] = struct{}{}
		}
		fmt.Println(line.String())
	}
	return nil
}
