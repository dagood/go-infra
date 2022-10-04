List of builds:

{{ range $version, $reportsByPipeline := .Reports }}

## {{ $version }}

    {{ range $pipeline, $reportsByBuildID := $reportsByPipeline }}

### {{ $pipeline }}

Build ID | Status | Started | Last Report
--- | --- | --- | ---
        {{- range $buildID, $value := $reportsByBuildID }}
{{ $buildID }} | {{ $value.Status }} | {{ $value.StartTime }} | {{ $value.LastUpdate }}
        {{ end }}
    {{ end }}
{{ end }}
