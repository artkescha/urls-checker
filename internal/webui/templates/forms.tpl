{{define "loginform"}}
    <div class="login">
      {{if .errmsg}}<p class="error">{{.errmsg}}</p>{{end}}
      <form method="post" action="/login">
        <ul class="wrapper">
          <li class="form-row">
            <label for="username">Username</label>
            <input type="text" id="username" name="username">
          </li>
          <li class="form-row">
            <label for="password">Password</label>
            <input type="password" id="password" name="password">
          </li>
          <li class="form-row">
            <button type="submit">Sign&nbsp;In</button>
          </li>
        </ul>
      </form>
    </div>
{{end}}

{{define "jobform"}}
      <section class="settings">
        <h2>File upload</h2>
{{range $key, $value := .forms}}{{template "uploadform" $value}}{{end}}
      </section>
      <section class="settings">
        <h2>Job Parameters</h2>
        <form  action="/job" method="post">
          <p><input type="checkbox" name="ignorecase" id="ignorecase" value="1" {{if .ignorecase}}checked{{end}} /> <label for="ignorecase">Ignore case of keywords</label></p>
          <h3>Protocols</h3>
          <p><input type="checkbox" name="http" id="http" value="1" {{if .http}}checked{{end}} /> <label for="http">HTTP</label></p>
          <p><input type="checkbox" name="https" id="https" value="1" {{if .https}}checked{{end}} /> <label for="https">HTTPS</label></p>
          <h3>Data</h3>
          <p><label for="skiplines">Skip the lines of the domains file:</label></p>
          <p><input type="text" name="skiplines" id="skiplines" value="{{.skiplines}}" /></p>
          <p><input type="checkbox" name="appendres" id="appendres" value="1" {{if .appendres}}checked{{end}} /> <label for="appendres">Complete the result file</label></p>
          <p><button type="submit">Run</button></p>
        </form>
      </section>
{{end}}

{{define "settingsform"}}
<form method="post" action="/settings">
<section class="settings">
<h2>Program settings</h2>
<p><label for="DomainResendCount">Number of redirects of erroneous domains to another worker</label></p>
<p><input type="number" name="DomainResendCount" id="DomainResendCount" value="{{.DomainResendCount}}" /></p>
<p><label for="WorkerPollInterval">Interval between polling workers</label></p>
<p><input type="text" name="WorkerPollInterval" id="WorkerPollInterval" value="{{.WorkerPollInterval.D}}" /></p>
<p><label for="WorkerRetryCount">Number of failed attempts to connect to the worker before deactivating it</label></p>
<p><input type="number" name="WorkerRetryCount" id="WorkerRetryCount" value="{{.WorkerRetryCount}}" /></p>
<p><label for="WorkerRecheckInterval">Interval for checking a deactivated worker</label></p>
<p><input type="text" name="WorkerRecheckInterval" id="WorkerRecheckInterval" value="{{.WorkerRecheckInterval.D}}" /></p>
</section>

<section class="settings">
<h2>Default worker settings</h2>
{{template "workerpar" .WorkerDefaults}}
</section>
<button type="submit">Submit</button>
</form>
{{end}}

{{define "workersform"}}
      <table class="workers">
        <tr class="workers">
          <th class="workers">Address</th>
          <th>Threads</th>
          <th>Timeout</th>
          <th>Job length</th>
          <th>&nbsp;</th>
        </tr>
{{range $key, $value := .workers}}        <tr>
          <td>{{$key}}</td>
          <td>{{filterDefault $value.Threads}}</td>
          <td>{{filterDefault $value.Timeout.D}}</td>
          <td>{{filterDefault $value.JobLen}}</td>
          <td>
            <a href="/worker?action=edit&address={{$key | urlescaper}}">Edit</a>
            <a href="/worker?action=delete&address={{$key | urlescaper}}">Delete</a>
          </td>
        </tr>{{end}}
      </table>
      <p><a href="/worker?action=add">Add</a></p>
{{end}}

{{define "workerform"}}
<form method="post" action="{{.backurl}}">
<section class="settings">
<input type="hidden" name="oldaddress" id="oldaddress" value="{{.oldaddress}}" />
<p><label for="address">Worker URL</label></p>
<p><input type="text" name="address" id="address" value="{{.address}}" /></p>
</section>
<section class="settings">
<p>To use the default values, leave the corresponding fields blank</p>
{{template "workerpar" .worker}}
</section>
<button type="submit">Submit</button>
</form>
{{end}}

{{define "workerpar"}}
<p><label for="login">Login and password to access the worker</label></p>
<p><input type="text" name="login" id="login" value="{{.Login}}" /></p>
<p><input type="password" name="password" id="password" value="{{.Password}}" /></p>
<p><label for="threads">Number of Polling Threads</label></p>
<p><input type="number" name="threads" id="threads" value="{{filterDefault .Threads}}" /></p>
<p><label for="timeout">Request timeout</label></p>
<p><input type="text" name="timeout" id="timeout" value="{{filterDefault .Timeout.D}}" /></p>
<p><label for="job_len">Number of domains in the job</label></p>
<p><input type="number" name="job_len" id="job_len" value="{{filterDefault .JobLen}}" /></p>
<p><label for="error_delay">Delay between retries in case of errors</label></p>
<p><input type="text" name="error_delay" id="error_delay" value="{{filterDefault .ErrorDelay.D}}" /></p>

<h3>Number of retries on errors:</h3>
<p><label for="dns">DNS errors (timeout, host not found)</label></p>
<p><input type="number" name="dns" id="dns" value="{{filterDefault .ErrorRetry.DNS}}" /></p>
<p><label for="connect">Connection errors (timeout, refused)</label></p>
<p><input type="number" name="connect" id="connect" value="{{filterDefault .ErrorRetry.Connect}}" /></p>
<p><label for="https">HTTPS errors (invalid certificates or algorithms)</label></p>
<p><input type="number" name="https" id="https" value="{{filterDefault .ErrorRetry.HTTPS}}" /></p>
<p><label for="http">HTTP errors (codes other than 20x)</label></p>
<p><input type="number" name="http" id="http" value="{{filterDefault .ErrorRetry.HTTP}}" /></p>
<p><label for="unknown">Other errors</label></p>
<p><input type="number" name="unknown" id="unknown" value="{{filterDefault .ErrorRetry.Unknown}}" /></p>
{{end}}

{{define "uploadform"}}
        <form enctype="multipart/form-data" action="/upload/{{.Path}}" method="post">
          <h3>{{.Message}}</h3>
          <p>
            <input type="file" name="{{.Path}}" id="{{.Path}}" />
            <button type="submit">Upload zip archive</button>
          </p>
{{if .ErrMsg}}
          <p class="error">{{.ErrMsg}}</p>
{{else}}
          <p>{{if .Loaded}}File downloaded, {{.Size}} bytes, {{.Lines}} lines. <a href="/download/{{.Path}}">Download</a>{{else}}File not uploaded{{end}}</p>
{{end}}
        </form>
{{end}}
