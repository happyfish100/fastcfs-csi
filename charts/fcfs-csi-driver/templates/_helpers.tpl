{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "fcfs-csi-driver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fcfs-csi-driver.fullname" -}}
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
Create chart name and version as used by the chart label.
*/}}
{{- define "fcfs-csi-driver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "fcfs-csi-driver.labels" -}}
{{ include "fcfs-csi-driver.selectorLabels" . }}
{{- end -}}

{{/*
Common selector labels
*/}}
{{- define "fcfs-csi-driver.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fcfs-csi-driver.name" . }}
{{- end -}}

{{/*
Convert the `--extra-volume-tags` command line arg from a map.
*/}}
{{- define "fcfs-csi-driver.extra-volume-tags" -}}
{{- $evt := default .Values.extraVolumeTags .Values.controller.extraVolumeTags }}
{{- $result := dict "pairs" (list) -}}
{{- range $key, $value := $evt -}}
{{- $noop := printf "%s=%s" $key $value | append $result.pairs | set $result "pairs" -}}
{{- end -}}
{{- if gt (len $result.pairs) 0 -}}
{{- printf "%s=%s" "- --extra-volume-tags" (join "," $result.pairs) -}}
{{- end -}}
{{- end -}}

{{/*
Handle http proxy env vars
*/}}
{{- define "fcfs-csi-driver.http-proxy" -}}
- name: HTTP_PROXY
  value: {{ .Values.proxy.http_proxy | quote }}
- name: HTTPS_PROXY
  value: {{ .Values.proxy.http_proxy | quote }}
- name: NO_PROXY
  value: {{ .Values.proxy.no_proxy | quote }}
{{- end -}}
