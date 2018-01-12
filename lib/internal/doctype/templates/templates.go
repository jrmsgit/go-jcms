package templates

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"

	"github.com/jrmsdev/go-jcms/lib/internal/doctype"
	"github.com/jrmsdev/go-jcms/lib/internal/doctype/base"
	"github.com/jrmsdev/go-jcms/lib/internal/doctype/templates/funcs"
	"github.com/jrmsdev/go-jcms/lib/internal/fsutils"
	"github.com/jrmsdev/go-jcms/lib/internal/logger"
	"github.com/jrmsdev/go-jcms/lib/internal/request"
	"github.com/jrmsdev/go-jcms/lib/internal/response"
	"github.com/jrmsdev/go-jcms/lib/internal/settings"
)

var log = logger.New("doctype.templates")

func init() {
	doctype.Register("templates", newEngine())
}

type engine struct {
	base.Engine
}

func newEngine() *engine {
	return &engine{base.New("templates")}
}

func (e *engine) Handle(
	ctx context.Context,
	resp *response.Response,
	req *request.Request,
	cfg *settings.Reader,
	docroot string,
) context.Context {
	var (
		ok        bool
		maintplfn string
		viewtplfn string
	)
	log.D("docroot %s", docroot)
	// get template files
	maintplfn, ok = getMainTpl(cfg, docroot, "main")
	if !ok {
		log.E("main template not found: %s", maintplfn)
		return resp.SetError(ctx, http.StatusInternalServerError,
			"main template not found")
	}
	viewtplfn, ok = getViewTpl(cfg, req, docroot)
	if !ok {
		log.E("view template not found: %s", viewtplfn)
		return resp.SetError(ctx, http.StatusNotFound, "not found")
	}
	// templates data
	tpldata := newData()
	return tplHandle(ctx, resp, req, cfg, docroot,
		maintplfn, viewtplfn, tpldata)
}

func (e *engine) HandleError(
	ctx context.Context,
	resp *response.Response,
	req *request.Request,
	cfg *settings.Reader,
	docroot string,
) context.Context {
	// get error template
	maintplfn, ok := getMainTpl(cfg, docroot, "error")
	if !ok {
		log.E("error template not found: %s", maintplfn)
		return resp.SetError(ctx, http.StatusInternalServerError,
			"error template not found")
	}
	// templates data
	tpldata := newErrorData()
	return tplHandle(ctx, resp, req, cfg, docroot, maintplfn, "", tpldata)
}

func tplHandle(
	ctx context.Context,
	resp *response.Response,
	_ *request.Request,
	_ *settings.Reader,
	docroot string,
	maintplfn string,
	viewtplfn string,
	tpldata *Data,
) context.Context {
	var (
		err     error
		tplname string
		maintpl *template.Template
		viewtpl *template.Template
	)
	// parse main template
	maintpl, err = parseMainTpl(maintplfn)
	if err != nil {
		log.E("parse main template: %s", err.Error())
		return resp.SetError(ctx, http.StatusInternalServerError,
			"ERROR: parse main template")
	}
	// parse view template (if provided)
	if viewtplfn != "" {
		viewtpl, err = parseViewTpl(maintpl, viewtplfn)
		if err != nil {
			log.E("parse view template: %s", err.Error())
			return resp.SetError(ctx, http.StatusInternalServerError,
				"ERROR: parse view template")
		}
	}
	// execute main template
	tplname = tplName(docroot, maintplfn)
	resp.SetTemplateLayout(tplname)
	log.D("exec main", tplname)
	err = execTpl(resp, maintpl, tpldata)
	if err != nil {
		log.E("exec main template: %s", err.Error())
		return resp.SetError(ctx, http.StatusInternalServerError,
			"ERROR: exec main template")
	}
	// execute view template (if provided)
	if viewtplfn != "" {
		tplname = tplName(docroot, viewtplfn)
		resp.SetTemplate(tplname)
		log.D("exec view", tplname)
		err = execTpl(resp, viewtpl, tpldata)
		if err != nil {
			log.E("exec view template: %s", err.Error())
			return resp.SetError(ctx, http.StatusInternalServerError,
				"ERROR: exec view template")
		}
	}
	resp.SetStatus(http.StatusOK)
	return ctx
}

func getMainTpl(cfg *settings.Reader, docroot, layout string) (string, bool) {
	filename := filepath.Join(docroot, layout+".tpl")
	if !fsutils.FileExists(filename) {
		return filename, false
	}
	return filename, true
}

func getViewTpl(
	cfg *settings.Reader,
	req *request.Request,
	docroot string,
) (string, bool) {
	fn := req.URL.Path
	if fn == "" || fn == "/" {
		fn = path.Clean(cfg.View.Path)
	}
	if fn == "" || fn == "/" {
		fn = "index"
	}
	filename := filepath.Join(docroot, fn+".html")
	if !fsutils.FileExists(filename) {
		return filename, false
	}
	return filename, true
}

func parseMainTpl(fn string) (*template.Template, error) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return template.New("maintpl").Funcs(funcs.Map).Parse(string(content))
}

func parseViewTpl(main *template.Template, fn string) (*template.Template, error) {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return template.Must(main.Clone()).Parse(string(content))
}

func tplName(docroot, filename string) string {
	n, err := filepath.Rel(docroot, filename)
	if err != nil {
		n = "ERROR:" + err.Error()
	}
	return n
}

func execTpl(
	resp *response.Response,
	tpl *template.Template,
	data *Data,
) error {
	return tpl.Execute(resp, data)
}
