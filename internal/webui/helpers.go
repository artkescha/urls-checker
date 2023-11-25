package webui

import (
	"errors"
	cw "example.com/scaner/internal/centerweb"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type (
	UploadFormData struct {
		Path    string
		Message string
		Loaded  bool
		Size    int64
		Lines   int
		ErrMsg  string
	}
)

var (
	nav = [][]string{
		{PATH_ROOT, TITLE_ROOT},
		{PATH_JOB, TITLE_JOB},
		{PATH_SETTINGS, TITLE_SETTINGS},
		{PATH_WORKERS, TITLE_WORKERS},
		{PATH_LOGOUT, TITLE_LOGOUT},
	}

	ufmessages = [][]string{
		{cw.FILE_DOMAINS, "Domains File"},
		{cw.FILE_LINKS, "Links file"},
		{cw.FILE_KEYWORDS, "Keywords file"},
	}

	statusmessages = map[int]string{
		cw.JOB_STATUS_IDLE:      "Job not active",
		cw.JOB_STATUS_STARTING:  "Job starts",
		cw.JOB_STATUS_WORK:      "Job in progress",
		cw.JOB_STATUS_STOPPING:  "Job stopping",
		cw.JOB_STATUS_COMPLETED: "Job completed",
		cw.JOB_STATUS_ERROR:     "Job error",
	}
)

func tplparams(title string) map[string]interface{} {
	return map[string]interface{}{
		"nav":   nav,
		"title": title,
	}
}

func checkWorkerParams(req *http.Request, dest *cw.WorkerConfig, errFields *[]string) {
	checkInt(req, "threads", &dest.Threads, errFields, -1, 10000, -1)
	checkDuration(req, "timeout", &dest.Timeout, errFields, -1, 10*time.Minute, -1)
	checkInt(req, "job_len", &dest.JobLen, errFields, -1, 100000, -1)
	checkDuration(req, "error_delay", &dest.ErrorDelay, errFields, -1, time.Minute, -1)
	checkInt(req, "dns", &dest.ErrorRetry.DNS, errFields, -1, 100, -1)
	checkInt(req, "connect", &dest.ErrorRetry.Connect, errFields, -1, 100, -1)
	checkInt(req, "https", &dest.ErrorRetry.HTTPS, errFields, -1, 100, -1)
	checkInt(req, "http", &dest.ErrorRetry.HTTP, errFields, -1, 100, -1)
	checkInt(req, "unknown", &dest.ErrorRetry.Unknown, errFields, -1, 100, -1)
}

func checkURL(addr string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("invalid URL " + addr + " : unsupported scheme")
	}
	if u.Hostname() == "" {
		return errors.New("invalid URL " + addr + " : empty hostname")
	}
	return nil
}

func checkInt(req *http.Request, name string, dest *int, errFields *[]string, min int, max int, emptyval int) {
	var (
		value int
		err   error
	)
	str := req.PostFormValue(name)
	if str == "" {
		value = emptyval
	} else {
		value, err = strconv.Atoi(str)
		if err != nil {
			*errFields = append(*errFields, name)
			return
		}
	}
	if value < min || value > max {
		*errFields = append(*errFields, name)
		return
	}
	*dest = value
}

func checkDuration(req *http.Request, name string, dest *cw.Duration, errFields *[]string, min time.Duration, max time.Duration, emptyval time.Duration) {
	var (
		value time.Duration
		err   error
	)
	str := req.PostFormValue(name)
	if str == "" {
		value = emptyval
	} else {
		value, err = time.ParseDuration(str)
		if err != nil {
			*errFields = append(*errFields, name)
			return
		}
	}
	if value < min || value > max {
		*errFields = append(*errFields, name)
		return
	}
	dest.D = value
}

func filterDefault(val interface{}) interface{} {
	switch v := val.(type) {
	case int:
		if v == -1 {
			return ""
		}

	case time.Duration:
		if v == -1 {
			return ""
		}

	case cw.Duration:
		if v.D == -1 {
			return ""
		}
	}
	return val
}

func formatErrMap(x map[string]int) template.HTML {
	r := ""
	keys := make([]string, 0, len(x))
	for k := range x {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		r += fmt.Sprintf("%s: %d<br />", k, x[k])
	}
	return template.HTML(r)
}

func uploadError(resp http.ResponseWriter, filename string, err error) {
	cw.JobStatus.Lock()
	fs := cw.JobStatus.FileStatus[filename]
	fs.Loaded = false
	fs.ErrMsg = err.Error()
	cw.JobStatus.FileStatus[filename] = fs
	cw.JobStatus.Unlock()
	log.Printf("File '%s' upload error: %s", filename, err.Error())
	//	http.Error(resp, err.Error(), http.StatusInternalServerError)
}
