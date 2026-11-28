{{/*
Standard Kubernetes labels. Called with a dict: (dict "root" $ "name" $name).
*/}}
{{- define "todo-platform.labels" -}}
{{ include "todo-platform.selectorLabels" . }}
app.kubernetes.io/part-of: todo-platform
app.kubernetes.io/managed-by: {{ .root.Release.Service }}
helm.sh/chart: {{ .root.Chart.Name }}-{{ .root.Chart.Version }}
{{- end }}

{{/*
Selector labels: must never change after the first install, because a
Deployment's selector is immutable.
*/}}
{{- define "todo-platform.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
{{- end }}

{{/* Name of the Secret holding the database password. */}}
{{- define "todo-platform.dbSecret" -}}
{{- .Values.database.existingSecret | default "db-credentials" }}
{{- end }}
