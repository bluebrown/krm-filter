package azurevaultsecrets

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/bluebrown/treasure-map/textfunc"
)

var TemplateRenderer = tpl()

// initialize a new template with all the good stuff as funcMap
func tpl() *template.Template {
	t := template.New("")
	fm := sprig.TxtFuncMap()
	fm["envToYaml"] = envFileToYaml
	fm["env"] = func(v any) any { panic("cannot use env") }
	fm["expandenv"] = func(v any) any { panic("cannot use expandenv") }
	t = t.Funcs(textfunc.MapClosure(fm, t))
	return t
}

// this function helps to use multi line vault secrets
// it replaces the first '=' with ': ' to make it valid yaml
// afterwards yaml.Unmarshal can be used on the template
func envFileToYaml(contents string) string {
	lines := strings.Split(contents, "\n")
	var op string
	for _, l := range lines {
		if l == "" {
			continue
		}
		op = fmt.Sprintf("%s\n%s\n", op, strings.Replace(l, "=", ": ", 1))
	}
	return op
}
