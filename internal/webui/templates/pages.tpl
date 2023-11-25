{{define "login"}}
{{template "prolog" .}}
{{template "loginform" .}}
{{template "epilog"}}
{{end}}

{{define "status"}}
{{template "prolog_full" .}}
{{template "statusform" .}}
{{template "epilog_full" .}}
{{end}}

{{define "job"}}
{{template "prolog_full" .}}
{{template "jobform" .}}
{{template "epilog_full" .}}
{{end}}

{{define "settings"}}
{{template "prolog_full" .}}
{{template "settingsform" .cfg}}
{{template "epilog_full" .}}
{{end}}

{{define "workers"}}
{{template "prolog_full" .}}
{{template "workersform" .}}
{{template "epilog_full" .}}
{{end}}

{{define "worker"}}
{{template "prolog_full" .}}
{{template "workerform" .}}
{{template "epilog_full" .}}
{{end}}
