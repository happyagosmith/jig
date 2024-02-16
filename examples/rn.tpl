# RELEASE NOTE
This is an example of Release Note that uses the generated valuse by jig

# DETTAGLI RILASCIO

|**Nome servizio**             | **versione** |
|------------------------------|--------------|
{{- range .services }}
{{- if and .label (ne .previousVersion .version) }}
|{{ .label }}|{{ .version }}|
{{- end }}
{{- end }}

# NUOVA BASELINE
|**Nome servizio**             | **versione** |
|------------------------------|--------------|
{{- range .services }}
{{- if .label }}
|{{ .label }}|{{ .version }}|
{{- end }}
{{- end }}

## NUOVE FUNZIONALITA'
{{- range  $key, $value := .generatedValues.features }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }}

## BUG RISOLTI
{{- range $key, $value := .generatedValues.bugs }}
{{- range  $value }}
- |{{ $key }}| {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }} 

## PROBLEMI NOTI
{{- range $key, $value := .generatedValues.knownIssues }}
{{- range  $value }}
- {{ .issueSummary }} [{{ .issueKey }}](https://happyagosmith.atlassian.net/browse/{{ .issueKey }})
{{- end }}
{{- end }}