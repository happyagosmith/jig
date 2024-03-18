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
- |{{ $key }}| {{ .issueSummary }} {{ .issueKey }}
{{- end }}
{{- end }}


or 

{{ range (issuesFlatList .generatedValues.features)}} 
{{- .issueKey}}: {{ .impactedService | join ","}}
{{end }}

## FIXED BUGs
{{- range $key, $value := .generatedValues.bugs }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} {{ .issueKey }}
{{- end }}
{{- end }} 


or 

{{range (issuesFlatList .generatedValues.bugs)}} 
{{ .issueKey}}: {{ .impactedService | join ","}}
{{end}}


## BREAKING CHANGES
{{- range $key, $value := .generatedValues.breakingChange }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} {{ .issueKey }}
{{- end }}
{{- end }} 

or 

{{range (issuesFlatList .generatedValues.breakingChange)}} 
{{ .issueKey}}: {{ .impactedService | join ","}}
{{end}}