{{- define "kubecoder.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kubecoder.fullname" -}}
{{- if .Values.fullnameOverride }}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{ else }}{{ printf "%s-%s" .Release.Name (include "kubecoder.name" .) | trunc 63 | trimSuffix "-" }}{{ end }}
{{- end }}

{{- define "kubecoder.postgresqlName" -}}
{{- printf "%s-postgresql" (include "kubecoder.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kubecoder.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "kubecoder.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "kubecoder.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubecoder.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "kubecoder.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{- define "kubecoder.databaseSecretName" -}}
{{- if .Values.postgresql.enabled -}}
{{- default (include "kubecoder.postgresqlName" .) .Values.postgresql.auth.existingSecret -}}
{{- else -}}
{{- required "externalDatabase.dsnSecretName is required when postgresql.enabled=false" .Values.externalDatabase.dsnSecretName -}}
{{- end -}}
{{- end }}

{{- define "kubecoder.databaseDSN" -}}
{{- printf "postgres://%s:%s@%s:5432/%s?sslmode=disable" (urlquery .Values.postgresql.auth.username) (urlquery .Values.postgresql.auth.password) (include "kubecoder.postgresqlName" .) (urlquery .Values.postgresql.auth.database) -}}
{{- end }}
