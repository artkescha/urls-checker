{{define "prolog"}}
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>{{.title}}</title>
    <link rel="stylesheet" href="/static/index.css">
  </head>
  <body>
{{end}}

{{define "epilog"}}
  </body>
</html>
{{end}}

{{define "prolog_full"}}
{{template "prolog" .}}
    <main>
      <nav>
        <ul>
{{range $key, $value := .nav}}
          <li><a href="{{index $value 0}}">{{index $value 1}}</a></li>
{{end}}
        </ul>
      </nav>
      <h1>{{.title}}</h1>
      {{if .errmsg}}<p class="error">{{.errmsg}}</p>{{end}}
{{end}}

{{define "epilog_full"}}
    </main>
{{if .errors}}
<script>
{{range $key, $value := .errors}}
document.getElementById("{{$value}}").className="error";
{{end}}
</script>
{{end}}
{{template "epilog"}}
{{end}}

