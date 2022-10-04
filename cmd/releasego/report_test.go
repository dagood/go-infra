package main

import (
	"path/filepath"
	"testing"
	"time"
)

func Test_reportComment_body(t *testing.T) {
	tests := []struct {
		name string
		args []buildReportPresentation
	}{
		{"empty", []buildReportPresentation{{BuildID: "123"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reportComment{
				before:  "Text before the report.",
				after:   "Text after the report.",
				reports: nil,
			}
			got, err := r.body(r)
			if err != nil {
				t.Errorf("(r *reportComment) body() error = %v", err)
				return
			}
			checkGolden(t, filepath.Join("testdata", "report", tt.name+".golden.md"), got)
		})
	}
}

func simpleReport(version, id, pipeline, symbol string) buildReportPresentation {
	return buildReportPresentation{
		buildReport: buildReport{
			BuildID:         "",
			BuildPipeline:   "",
			BuildCollection: "",
			BuildProject:    "",
			Status:          "",
			Version:         "",
			LastUpdate:      time.Time{},
			StartTime:       time.Time{},
		},
		BuildURL:    "",
		BuildSymbol: "",
	}
}
