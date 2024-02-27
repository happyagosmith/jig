# RELEASE NOTE
This is an example of Release Note that uses the generated valuse by jig

# RELEASE DETAILS

|**Nome servizio**             | **versione** |
|------------------------------|--------------|
{{- range .services }}
{{- if and .label (ne .previousVersion .version) }}
|{{ .label }}|{{ .version }}|
{{- end }}
{{- end }}

# NEW BASELINE
|**Nome servizio**             | **versione** |
|------------------------------|--------------|
{{- range .services }}
{{- if .label }}
|{{ .label }}|{{ .version }}|
{{- end }}
{{- end }}

## NEW FEATUREs
{{- range  $key, $value := .generatedValues.features }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }}

## FIXED BUGs
{{- range $key, $value := .generatedValues.bugs }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }} 

## KNOWN ISSUEs
{{- range $key, $value := .generatedValues.knownIssues }}
{{- range  $value }}
- {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }}