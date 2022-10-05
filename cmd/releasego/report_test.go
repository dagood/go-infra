package main

import (
	"path/filepath"
	"testing"
)

func Test_reportComment_body(t *testing.T) {
	r := func(version, id, pipeline, status string) buildReportPresentation {
		b := buildReport{
			Version:         version,
			BuildID:         id,
			BuildPipeline:   pipeline,
			BuildCollection: "https://example.org/",
			BuildProject:    "public",
			Status:          status,
		}
		return b.present()
	}

	tests := []struct {
		name    string
		reports []buildReportPresentation
	}{
		{"realistic", []buildReportPresentation{
			r("1.18.2-1", "1234", "microsoft-go-infra-release-build", "Succeeded"),
			r("1.18.2-1", "1238", "microsoft-go-infra-release-build", "InProgress"),
			r("1.18.2-1", "1500", "microsoft-go-infra-release-go-images", "InProgress"),
			r("1.19.1-1", "1900", "microsoft-go-infra-release-build", "NotStarted"),
			r("1.18.2-1-fips", "1239", "microsoft-go-infra-release-build", "Failed"),
			r("1.18.2-1", "1233", "microsoft-go-infra-release-build", "Failed"),
			r("1.18.2-1", "1300", "microsoft-go-infra-release-build", "NotStarted"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reportComment{
				before:  "Text before the report.\n\n",
				after:   "\n\nText after the report.",
				reports: tt.reports,
			}
			got, err := r.body()
			if err != nil {
				t.Errorf("(r *reportComment) body() error = %v", err)
				return
			}
			checkGolden(t, "Test_reportComment_body", filepath.Join("testdata", "report", tt.name+".golden.md"), got)
		})
	}
}
