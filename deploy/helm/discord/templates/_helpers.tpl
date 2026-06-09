{{/* Base name, overridable. */}}
{{- define "discord.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Fully qualified app name. */}}
{{- define "discord.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "discord.name" . -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/* Common labels shared by every object. */}}
{{- define "discord.labels" -}}
app.kubernetes.io/name: {{ include "discord.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: discord
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{- end -}}

{{/* Backend selector labels. */}}
{{- define "discord.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "discord.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: backend
{{- end -}}

{{/* Frontend selector labels. */}}
{{- define "discord.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "discord.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: frontend
{{- end -}}

{{/* Backend / frontend resource names. */}}
{{- define "discord.backend.fullname" -}}
{{- printf "%s-backend" (include "discord.fullname" .) -}}
{{- end -}}
{{- define "discord.frontend.fullname" -}}
{{- printf "%s-frontend" (include "discord.fullname" .) -}}
{{- end -}}

{{/*
ALLOWED_ORIGINS value.
If backend.env.allowedOrigins is set explicitly use it as-is.
Otherwise derive from ingress.host + TLS flag so there is one source of truth.
*/}}
{{- define "discord.allowedOrigins" -}}
{{- if .Values.backend.env.allowedOrigins -}}
{{- .Values.backend.env.allowedOrigins -}}
{{- else -}}
{{- printf "%s://%s" (ternary "https" "http" .Values.ingress.tls.enabled) .Values.ingress.host -}}
{{- end -}}
{{- end -}}

{{/* Resolved image references (tag defaults to appVersion). */}}
{{- define "discord.backend.image" -}}
{{- $tag := .Values.backend.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" .Values.backend.image.registry .Values.backend.image.repository $tag -}}
{{- end -}}
{{- define "discord.frontend.image" -}}
{{- $tag := .Values.frontend.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" .Values.frontend.image.registry .Values.frontend.image.repository $tag -}}
{{- end -}}
