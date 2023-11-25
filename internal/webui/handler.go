package webui

import (
	"archive/zip"
	"bufio"
	"errors"
	cw "example.com/scaner/internal/centerweb"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	COOKIE_NAME = "fscc_"

	PATH_ROOT     = "/"
	PATH_JOB      = "/job"
	PATH_SETTINGS = "/settings"
	PATH_WORKERS  = "/workers"
	PATH_WORKER   = "/worker"
	PATH_LOGIN    = "/login"
	PATH_LOGOUT   = "/logout"
	PATH_UPLOAD   = "/upload/"
	PATH_DOWNLOAD = "/download/"

	TITLE_ROOT     = "Status"
	TITLE_JOB      = "Job"
	TITLE_SETTINGS = "Settings"
	TITLE_WORKERS  = "Workers"
	TITLE_WORKER   = "Worker"
	TITLE_LOGOUT   = "Logout"
)

type (
	Handler struct {
		cstore    *sessions.CookieStore
		templates *template.Template
		username  string
		password  string
		muxFileOp sync.Mutex
		curJob    *cw.Job
	}
)

func (this *Handler) InitSessions() {
	this.cstore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	this.cstore.Options.HttpOnly = true
}

func (this *Handler) InitTemplates() (err error) {
	funcMap := template.FuncMap{
		"urlescaper":    template.URLQueryEscaper,
		"filterDefault": filterDefault,
		"formatErrMap":  formatErrMap,
		"formatFloat":   strconv.FormatFloat,
		"add":           func(x, y int) int { return x + y },
	}
	this.templates, err = template.New("").Funcs(funcMap).ParseFS(templatesContent, "templates/*")
	return
}

func (this *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	var auth_ok bool
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.URL.Path)
	session, _ := this.cstore.Get(req, COOKIE_NAME)
	if tmp, ok := session.Values["auth_ok"]; ok {
		auth_ok = tmp.(bool)
	} else {
		auth_ok = false
	}
	session.Values["auth_ok"] = auth_ok
	session.Save(req, resp)

	// если юзер не авторизован, отправим его логиниться.
	if !auth_ok && req.URL.Path != PATH_LOGIN {
		http.Redirect(resp, req, PATH_LOGIN, http.StatusSeeOther)
		return
	}

	switch {
	case req.URL.Path == PATH_ROOT:
		this.status(resp, req)

	case req.URL.Path == PATH_JOB:
		this.job(resp, req)

	case req.URL.Path == PATH_SETTINGS:
		this.settings(resp, req)

	case req.URL.Path == PATH_WORKERS:
		this.workers(resp, req)

	case req.URL.Path == PATH_WORKER:
		this.worker(resp, req)

	case req.URL.Path == PATH_LOGIN:
		this.login(session, resp, req)

	case req.URL.Path == PATH_LOGOUT:
		this.logout(session, resp, req)

	case strings.HasPrefix(req.URL.Path, PATH_UPLOAD):
		this.upload(resp, req)

	case strings.HasPrefix(req.URL.Path, PATH_DOWNLOAD):
		this.download(resp, req)

	default:
		http.Error(resp, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func (this *Handler) login(session *sessions.Session, resp http.ResponseWriter, req *http.Request) {
	errmsg := ""
	if req.Method == http.MethodPost {
		auth_ok := req.PostFormValue("username") == this.username && req.PostFormValue("password") == this.password
		session.Values["auth_ok"] = auth_ok
		session.Save(req, resp)
		if auth_ok {
			http.Redirect(resp, req, PATH_ROOT, http.StatusSeeOther)
			return
		} else {
			errmsg = "Invalid credentials"
		}
	}
	this.templates.ExecuteTemplate(resp, "login", map[string]string{"title": "Login", "errmsg": errmsg})
}

func (this *Handler) logout(session *sessions.Session, resp http.ResponseWriter, req *http.Request) {
	session.Values["auth_ok"] = false
	session.Save(req, resp)
	http.Redirect(resp, req, PATH_LOGIN, http.StatusSeeOther)
}

func (this *Handler) status(resp http.ResponseWriter, req *http.Request) {
	if req.FormValue("action") == "stop" {
		if this.curJob != nil && this.curJob.Status.GetStatus() == cw.JOB_STATUS_WORK {
			this.curJob.Stop(false)
			this.curJob = nil
		}
	}
	resp.Header().Set("Refresh", "10; url=/")
	cw.JobStatus.RLock()
	defer cw.JobStatus.RUnlock()
	//	log.Printf("%+v", cw.JobStatus)
	params := tplparams("Job completion status")
	params["statusmsg"] = statusmessages[int(cw.JobStatus.GetStatus())]
	if cw.JobStatus.LastError != nil {
		params["statuserr"] = cw.JobStatus.LastError.Error()
	}
	params["worktime"] = cw.JobStatus.EndTime.Sub(cw.JobStatus.StartTime).Round(time.Second).String()
	params["status"] = cw.JobStatus
	params["linesread"] = atomic.LoadUint64(&cw.JobStatus.LineNum)
	params["linestotal"] = cw.JobStatus.FileStatus[cw.FILE_DOMAINS].Lines
	this.templates.ExecuteTemplate(resp, "status", params)
}

func (this *Handler) job(resp http.ResponseWriter, req *http.Request) {
	params := tplparams("Job Options")
	if req.Method == http.MethodPost { // проверка данных формы
		errFields := []string{}
		status := cw.JobStatus.GetStatus()
		if status != cw.JOB_STATUS_IDLE && status != cw.JOB_STATUS_COMPLETED && status != cw.JOB_STATUS_ERROR {
			params["errmsg"] = "Job already started"
		} else {
			ignorecase := req.FormValue("ignorecase") != ""
			params["ignorecase"] = ignorecase
			use_http := req.FormValue("http") != ""
			params["http"] = use_http
			use_https := req.FormValue("https") != ""
			params["https"] = use_https
			skiplines := 0
			if !use_http && !use_https {
				errFields = append(errFields, "http", "https")
				params["errmsg"] = "No protocol is selected"
			}
			checkInt(req, "skiplines", &skiplines, &errFields, 0, 100000000, 0)
			params["skiplines"] = skiplines
			appendres := req.FormValue("appendres") != ""
			params["appendres"] = appendres

			if len(errFields) == 0 { // ошибок нет
				log.Printf("Job started")

				// здесь надо запустить задание
				this.curJob = cw.NewJob(&cw.WCfg, &cw.JobStatus)
				err := this.curJob.Start(cw.JobConfig{
					IgnoreCase: ignorecase,
					UseHTTP:    use_http,
					UseHTTPS:   use_https,
					SkipLines:  skiplines,
					AppendRes:  appendres,
				})
				if err != nil {
					params["errmsg"] = err.Error()
				} else {
					http.Redirect(resp, req, PATH_ROOT, http.StatusSeeOther) // все успешно, обновим страницу
				}

			} else {
				log.Printf("Error msg: %s fields: %+v", params["errmsg"], errFields)
				if params["errmsg"] == nil || params["errmsg"].(string) == "" {
					params["errmsg"] = "Data error"
				}
				params["errors"] = errFields
			}
		}
	}

	forms := make([]UploadFormData, len(ufmessages))
	cw.JobStatus.RLock()
	for i, s := range ufmessages {
		forms[i] = UploadFormData{
			Path:    s[0],
			Message: s[1],
			Loaded:  cw.JobStatus.FileStatus[s[0]].Loaded,
			Lines:   cw.JobStatus.FileStatus[s[0]].Lines,
			Size:    cw.JobStatus.FileStatus[s[0]].Size,
			ErrMsg:  cw.JobStatus.FileStatus[s[0]].ErrMsg,
		}
	}
	cw.JobStatus.RUnlock()
	params["forms"] = forms
	this.templates.ExecuteTemplate(resp, "job", params)
}

func (this *Handler) settings(resp http.ResponseWriter, req *http.Request) {
	cw.MuxWCfg.Lock() // чтобы не было гонки
	defer cw.MuxWCfg.Unlock()

	params := tplparams("Settings")
	if req.Method == http.MethodPost { // проверка данных формы
		newCfg := cw.WebConfig{}
		errFields := []string{}

		checkInt(req, "DomainResendCount", &newCfg.DomainResendCount, &errFields, 0, 100, -1)
		checkDuration(req, "WorkerPollInterval", &newCfg.WorkerPollInterval, &errFields, 100*time.Millisecond, 24*time.Hour, 0)
		checkInt(req, "WorkerRetryCount", &newCfg.WorkerRetryCount, &errFields, 0, 100, -1)
		checkDuration(req, "WorkerRecheckInterval", &newCfg.WorkerRecheckInterval, &errFields, time.Second, 24*time.Hour, 0)
		newCfg.WorkerDefaults.Login = req.PostFormValue("login")
		newCfg.WorkerDefaults.Password = req.PostFormValue("password")
		checkInt(req, "threads", &newCfg.WorkerDefaults.Threads, &errFields, 1, 10000, -1)
		checkDuration(req, "timeout", &newCfg.WorkerDefaults.Timeout, &errFields, time.Second, 10*time.Minute, 0)
		checkInt(req, "job_len", &newCfg.WorkerDefaults.JobLen, &errFields, 1, 100000, -1)
		checkDuration(req, "error_delay", &newCfg.WorkerDefaults.ErrorDelay, &errFields, time.Millisecond, time.Minute, 0)
		checkInt(req, "dns", &newCfg.WorkerDefaults.ErrorRetry.DNS, &errFields, 0, 100, 0)
		checkInt(req, "connect", &newCfg.WorkerDefaults.ErrorRetry.Connect, &errFields, 0, 100, -1)
		checkInt(req, "https", &newCfg.WorkerDefaults.ErrorRetry.HTTPS, &errFields, 0, 100, -1)
		checkInt(req, "http", &newCfg.WorkerDefaults.ErrorRetry.HTTP, &errFields, 0, 100, -1)
		checkInt(req, "unknown", &newCfg.WorkerDefaults.ErrorRetry.Unknown, &errFields, 0, 100, -1)

		if len(errFields) == 0 { // ошибок нет, сохраним настройки
			newCfg.Workers = cw.WCfg.Workers
			cw.WCfg = newCfg
			err := cw.SaveWebConfig()
			if err != nil {
				log.Printf("Settings save error: %s", err.Error())
				params["errmsg"] = err.Error()
			} else {
				http.Redirect(resp, req, PATH_SETTINGS, http.StatusSeeOther) // все успешно, обновим страницу
			}
		} else { // юзер что-то ввел неправильно
			params["errmsg"] = "Data error"
			params["errors"] = errFields
		}
		params["cfg"] = newCfg
	} else {
		params["cfg"] = cw.WCfg
	}
	this.templates.ExecuteTemplate(resp, "settings", params)
}

func (this *Handler) workers(resp http.ResponseWriter, req *http.Request) {
	cw.MuxWCfg.RLock() // чтобы не было гонки
	defer cw.MuxWCfg.RUnlock()
	params := tplparams("Workers")
	params["workers"] = cw.WCfg.Workers
	this.templates.ExecuteTemplate(resp, "workers", params)
}

func (this *Handler) worker(resp http.ResponseWriter, req *http.Request) {
	var (
		params map[string]interface{}
		neww   cw.WorkerConfig
		ok     bool
	)
	cw.MuxWCfg.Lock() // чтобы не было гонки
	defer cw.MuxWCfg.Unlock()

	reqAction := req.FormValue("action")
	reqAddress := req.FormValue("address")
	reqOldAddress := req.FormValue("oldaddress")
	if req.Method == http.MethodGet {
		a, err := url.QueryUnescape(reqAddress)
		if err == nil {
			reqAddress = a
		}
		reqOldAddress = reqAddress
	}

	switch reqAction {
	case "add":
		neww = cw.WorkerConfig{
			Login:      "",
			Password:   "",
			Threads:    -1,
			Timeout:    cw.Duration{D: -1},
			JobLen:     -1,
			ErrorDelay: cw.Duration{D: -1},
			ErrorRetry: cw.ErrorPolicy{
				DNS:     -1,
				Connect: -1,
				HTTPS:   -1,
				HTTP:    -1,
				Unknown: -1,
			},
		}
		params = tplparams("Add worker")

	case "edit":
		neww, ok = cw.WCfg.Workers[reqAddress]
		_, ok2 := cw.WCfg.Workers[reqOldAddress]
		if !ok && !ok2 {
			log.Printf("Worker not exists: %s", reqAddress)
			http.Redirect(resp, req, PATH_WORKERS, http.StatusSeeOther) // редактируем несуществующий воркер
			return
		}
		params = tplparams("Edit worker")

	case "delete":
		log.Printf("Delete worker %s", reqAddress)
		delete(cw.WCfg.Workers, reqAddress)
		err := cw.SaveWebConfig()
		if err != nil {
			log.Printf("Worker delete error: %s", err.Error())
		}
		http.Redirect(resp, req, PATH_WORKERS, http.StatusSeeOther) // возврат после удаления
		return

	default:
		log.Printf("Unknown worker action: %s", reqAction)
		http.Redirect(resp, req, PATH_WORKERS, http.StatusSeeOther) // неизвестная операция, возврат
		return
	}

	params["backurl"] = req.URL.String()
	params["errmsg"] = ""
	params["address"] = reqAddress
	params["oldaddress"] = reqOldAddress

	if req.Method == http.MethodPost { // проверка данных формы
		errFields := []string{}
		err := checkURL(reqAddress)
		if err != nil {
			params["errmsg"] = err.Error()
			errFields = append(errFields, "address")
		}
		neww.Login = req.PostFormValue("login")
		neww.Password = req.PostFormValue("password")
		checkWorkerParams(req, &neww, &errFields)

		if reqAction == "add" { // добавляем новый воркер
			log.Printf("Add worker")
			if _, ok = cw.WCfg.Workers[reqAddress]; ok {
				params["errmsg"] = "Address already exists"
				errFields = append(errFields, "address")
			}
		} else { // редактируем воркер
			log.Printf("Edit worker")
			if _, ok = cw.WCfg.Workers[reqOldAddress]; !ok {
				log.Printf("Worker not exists: %s", reqAddress)
				http.Redirect(resp, req, PATH_WORKERS, http.StatusSeeOther) // редактируем несуществующий воркер
				return
			} else if reqOldAddress != reqAddress {
				if _, ok = cw.WCfg.Workers[reqAddress]; ok { // воркер переименован в уже существующий
					params["errmsg"] = "Address already exists"
					errFields = append(errFields, "address")
				}
			}
		}

		log.Printf("Worker data: %+v", neww)
		if len(errFields) == 0 { // ошибок нет, сохраним настройки
			if reqOldAddress != reqAddress {
				delete(cw.WCfg.Workers, reqOldAddress)
			}
			cw.WCfg.Workers[reqAddress] = neww
			err := cw.SaveWebConfig()
			if err != nil {
				log.Printf("Settings save error: %s", err.Error())
				params["errmsg"] = err.Error()
			} else {
				log.Printf("Worker saved")
				http.Redirect(resp, req, PATH_WORKERS, http.StatusSeeOther) // все успешно, обновим страницу
			}
		} else {
			log.Printf("Error msg: %s fields: %+v", params["errmsg"], errFields)
			if params["errmsg"].(string) == "" {
				params["errmsg"] = "Data error"
			}
			params["errors"] = errFields
		}
	}

	params["worker"] = neww
	this.templates.ExecuteTemplate(resp, "worker", params)
}

func (this *Handler) upload(resp http.ResponseWriter, req *http.Request) {
	if cw.JobStatus.GetStatus() == cw.JOB_STATUS_WORK {
		http.Redirect(resp, req, PATH_ROOT, http.StatusSeeOther)
		return
	}
	if req.Method != http.MethodPost {
		http.Error(resp, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	this.muxFileOp.Lock()
	defer this.muxFileOp.Unlock()
	filename := req.URL.Path[len(PATH_UPLOAD):]
	if _, ok := cw.JobStatus.FileStatus[filename]; !ok {
		http.Error(resp, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	const maxFileSize = 500 * 1024 * 1024 // 500MB

	req.ParseMultipartForm(maxFileSize)
	upfile, header, err := req.FormFile(filename)
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			uploadError(resp, filename, err)
		}
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}
	defer upfile.Close()

	dst, err := os.CreateTemp(cw.Cfg.DataDir, "upload-*.zip")
	if err != nil {
		uploadError(resp, filename, err)
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}
	defer func() {
		dst.Close()
		os.Remove(dst.Name())
	}()

	_, err = io.Copy(dst, upfile)
	if err != nil {
		uploadError(resp, filename, err)
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}

	// открытие zip архива
	r, err := zip.OpenReader(dst.Name())
	if err != nil {
		uploadError(resp, filename, err)
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}
	defer r.Close()

	if len(r.File) == 0 {
		uploadError(resp, filename, fmt.Errorf("upload archive is empty"))
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}

	saveFile, err := os.OpenFile(filepath.Join(cw.Cfg.DataDir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		uploadError(resp, filename, err)
		http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
		return
	}
	defer saveFile.Close()

	lines := 0
	for _, f := range r.File {
		file, err := f.Open()
		if err != nil {
			uploadError(resp, filename, err)
			http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
			file.Close()
			return
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			_, err = saveFile.Write(scanner.Bytes())
			if err != nil {
				uploadError(resp, filename, err)
				http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
				file.Close()
				return
			}
			saveFile.Write([]byte("\n"))
			lines++
		}
		file.Close()
		err = scanner.Err()
		if err != nil {
			uploadError(resp, filename, err)
			http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther)
			return
		}
	}

	cw.JobStatus.Lock()
	fs := cw.JobStatus.FileStatus[filename]
	fs.Loaded = true
	fs.ErrMsg = ""
	fs.Lines = lines
	fs.Size = header.Size
	cw.JobStatus.FileStatus[filename] = fs
	cw.JobStatus.Save()
	cw.JobStatus.Unlock()
	log.Printf("File '%s' uploaded, %d lines, %d bytes", filename, lines, header.Size)
	http.Redirect(resp, req, PATH_JOB, http.StatusSeeOther) // все успешно, обновим страницу
}

func (this *Handler) download(resp http.ResponseWriter, req *http.Request) {
	this.muxFileOp.Lock()
	defer this.muxFileOp.Unlock()
	filename := req.URL.Path[len(PATH_DOWNLOAD):]
	if filename != cw.FILE_RESULTS && filename != cw.FILE_ERRORS {
		if s, ok := cw.JobStatus.FileStatus[filename]; !ok || !s.Loaded {
			http.Error(resp, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
	}

	infile, err := os.OpenFile(filepath.Join(cw.Cfg.DataDir, filename), os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("File '%s' download error: %s", filename, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}
	defer infile.Close()

	resp.Header().Set("Content-Type", "text/plain")
	resp.Header().Set("Content-Disposition", `attachment; filename="`+filename+`.txt"`)

	_, err = io.Copy(resp, infile)
	if err != nil {
		log.Printf("File '%s' download error: %s", filename, err.Error())
		return
	}
}
