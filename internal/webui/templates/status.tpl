{{define "statusform"}}
      <p><strong>{{.statusmsg}}</strong></p>
      {{if .statuserr}}<p class="error">{{.statuserr}}</p>{{end}}
      <p>Running time: {{.worktime}}</p>
      <p>Read lines of the input file: {{.linesread}} out of {{.linestotal}}</p>
      <h2>Results</h2>
      <table class="workers">
        <tr>
          <th>Domains processed</th>
          <th>Successfully</th>
          <th>With&nbsp;errors</th>
          <th>Total Requests</th>
          <th>With&nbsp;errors</th>
          <th>Requests per&nbsp;second</th>
          <th>Request errors</th>
        </tr>
        <tr>
          <td>{{add .status.Totals.Success .status.Totals.Failed}}</td>
          <td>{{.status.Totals.Success}}</td>
          <td>{{.status.Totals.Failed}}</td>
          <td>{{.status.Totals.Requests}}</td>
          <td>{{.status.Totals.ReqErrors}}</td>
          <td>{{formatFloat .status.Totals.Speed 'f' 2 64}}</td>
          <td>{{formatErrMap .status.Totals.ErrStat}}</td>
        </tr>
      </table>
      <p><a href="/download/results">Results</a></p>
      <p><a href="/download/errors">Error Log</a></p>
      <form method="get" action="/">
        <input type="hidden" name="action" id="action" value="stop" />
        <p><button type="submit">Stop working</button></p>
      </form>
      
      <h2>Workers stats</h2>
      <table class="workers">
        <tr>
          <th>Address</th>
          <th>Domains processed</th>
          <th>Successfully</th>
          <th>With&nbsp;errors</th>
          <th>Total Requests</th>
          <th>With&nbsp;errors</th>
          <th>Requests per&nbsp;second</th>
          <th>Request errors</th>
        </tr>
{{range $key, $value := .status.WorkerStatus}}        <tr>
          <td>{{$key}}</td>
          <td>{{add $value.Success $value.Failed}}</td>
          <td>{{$value.Success}}</td>
          <td>{{$value.Failed}}</td>
          <td>{{$value.Requests}}</td>
          <td>{{$value.ReqErrors}}</td>
          <td>{{formatFloat $value.Speed 'f' 2 64}}</td>
          <td>{{formatErrMap $value.ErrStat}}</td>
        </tr>{{end}}
      </table>
{{end}}
