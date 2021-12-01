package reconciler

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/valyala/fasttemplate"
)

var defaultRenderer = fastTemplateRenderer{}

type fastTemplateRenderer struct {
}

func (r fastTemplateRenderer) RenderTemplateParams(tmpl *kustomizev1.Kustomization, params map[string]string) (*kustomizev1.Kustomization, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("application template is empty ")
	}

	if len(params) == 0 {
		return tmpl, nil
	}

	tmplBytes, err := json.Marshal(tmpl)
	if err != nil {
		return nil, err
	}

	fstTmpl := fasttemplate.New(string(tmplBytes), "{{", "}}")
	replacedTmplStr, err := r.replace(fstTmpl, params, true)
	if err != nil {
		return nil, err
	}

	var replacedTmpl kustomizev1.Kustomization
	err = json.Unmarshal([]byte(replacedTmplStr), &replacedTmpl)
	if err != nil {
		return nil, err
	}
	return &replacedTmpl, nil
}

// Replace executes basic string substitution of a template with replacement values.
// 'allowUnresolved' indicates whether or not it is acceptable to have unresolved variables
// remaining in the substituted template.
func (r fastTemplateRenderer) replace(fstTmpl *fasttemplate.Template, replaceMap map[string]string, allowUnresolved bool) (string, error) {
	var unresolvedErr error
	replacedTmpl := fstTmpl.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {

		trimmedTag := strings.TrimSpace(tag)

		replacement, ok := replaceMap[trimmedTag]
		if len(trimmedTag) == 0 || !ok {
			if allowUnresolved {
				// just write the same string back
				return w.Write([]byte(fmt.Sprintf("{{%s}}", tag)))
			}
			unresolvedErr = fmt.Errorf("failed to resolve {{%s}}", tag)
			return 0, nil
		}
		// The following escapes any special characters (e.g. newlines, tabs, etc...)
		// in preparation for substitution
		replacement = strconv.Quote(replacement)
		replacement = replacement[1 : len(replacement)-1]
		return w.Write([]byte(replacement))
	})
	if unresolvedErr != nil {
		return "", unresolvedErr
	}

	return replacedTmpl, nil
}
