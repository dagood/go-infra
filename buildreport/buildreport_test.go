package buildreport

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/microsoft/go-infra/goldentest"
)

func Test_parseReportComment(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name string
		args args
		want commentBody
	}{
		{
			"no-section",
			args{"Comment body!"},
			commentBody{"Comment body!", "", nil}},
		{
			"no-data",
			args{"Before" + beginDataSectionMarker + "" + endDataSectionMarker + "After"},
			commentBody{"Before", "After", nil},
		},
		{
			"data",
			args{"Before" + beginDataSectionMarker + beginDataMarker + "[]" + endDataMarker + endDataSectionMarker + "After"},
			commentBody{"Before", "After", make([]State, 0)},
		},
		{
			"null",
			args{"Before" + beginDataSectionMarker + beginDataMarker + "null" + endDataMarker + endDataSectionMarker + "After"},
			commentBody{"Before", "After", nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseReportComment(tt.args.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseReportComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commentBody_body(t *testing.T) {
	exampleTime, err := time.Parse(time.RFC3339, "2012-03-28T01:02:03Z")
	if err != nil {
		t.Fatal(err)
	}

	r := func(version, id, pipeline, symbol string) State {
		b := State{
			Version:       version,
			BuildID:       id,
			BuildPipeline: pipeline,
			BuildURL:      "https://example.org/",
			BuildSymbol:   symbol,
			StartTime:     exampleTime,
			LastUpdate:    exampleTime.Add(time.Minute * 5),
		}
		return b
	}

	tests := []struct {
		name    string
		reports []State
	}{
		{
			"realistic",
			[]State{
				r("1.18.2-1", "1234", "microsoft-go-infra-release-build", SuccessSymbol),
				r("1.18.2-1", "1238", "microsoft-go-infra-release-build", InProgressSymbol),
				r("1.18.2-1", "1500", "microsoft-go-infra-release-go-images", InProgressSymbol),
				r("1.19.1-1", "1900", "microsoft-go-infra-release-build", NotStartedSymbol),
				r("1.18.2-1-fips", "1239", "microsoft-go-infra-release-build", FailedSymbol),
				r("1.18.2-1", "1233", "microsoft-go-infra-release-build", FailedSymbol),
				r("1.18.2-1", "1300", "microsoft-go-infra-release-build", NotStartedSymbol),
			},
		},
		{"none", nil},
		{
			"no version",
			[]State{r("", "1234", "microsoft-go-infra-start", InProgressSymbol)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := commentBody{
				before:  "Text before the report.\n\n",
				after:   "\n\nText after the report.",
				reports: tt.reports,
			}
			got, err := cb.body()
			if err != nil {
				t.Errorf("(r *reportComment) body() error = %v", err)
				return
			}
			goldentest.Check(t, "go test ./buildreport -run "+t.Name(), filepath.Join("testdata", "report", "body."+tt.name+".golden.md"), got)
		})
	}
}

func Test_commentBody_body_UpdateExisting(t *testing.T) {
	cb := commentBody{
		reports: []State{
			{
				Version:       "1.2.3",
				BuildPipeline: "microsoft-go",
				BuildID:       "1234",
				BuildSymbol:   InProgressSymbol,
			},
		},
	}
	cb.update(State{
		BuildID: "1234",
		// In practice, only the build symbol is expected to be updated, but test here that only the
		// BuildID is matched for a successful update.
		Version:       "1.2.4",
		BuildPipeline: "microsoft-go-changed-name",
		BuildSymbol:   SuccessSymbol,
	})
	got, err := cb.body()
	if err != nil {
		t.Errorf("(r *reportComment) body() error = %v", err)
		return
	}
	goldentest.Check(t, "go test ./buildreport -run "+t.Name(), filepath.Join("testdata", "report", "update-existing.golden.md"), got)
}
