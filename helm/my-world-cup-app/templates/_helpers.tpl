{{/*
Expand the name of the chart.
*/}}
{{- define "my-world-cup-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "my-world-cup-app.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Chart name and version label.
*/}}
{{- define "my-world-cup-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels.
*/}}
{{- define "my-world-cup-app.labels" -}}
helm.sh/chart: {{ include "my-world-cup-app.chart" . }}
{{ include "my-world-cup-app.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels.
*/}}
{{- define "my-world-cup-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "my-world-cup-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Service account name.
*/}}
{{- define "my-world-cup-app.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "my-world-cup-app.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Fail fast when rendered by anything older than Helm 4. The chart is only
tested against Helm 4 (see REQUIREMENTS.md); Helm 3 is no longer supported.
*/}}
{{- define "my-world-cup-app.requireHelm4" -}}
{{- if semverCompare "<4.0.0-0" .Capabilities.HelmVersion.Version -}}
{{- fail (printf "chart %s requires Helm 4 (found %s). Helm 3 is no longer supported - see https://helm.sh/docs/intro/install/" .Chart.Name .Capabilities.HelmVersion.Version) -}}
{{- end -}}
{{- end -}}
