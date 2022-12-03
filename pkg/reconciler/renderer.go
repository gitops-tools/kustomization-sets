package reconciler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/gitops-tools/pkg/sanitize"
)

var funcMap = template.FuncMap{
	"sanitize": sanitize.SanitizeDNSName,
}

func renderTemplateParams(tmpl *kustomizev1.Kustomization, params map[string]any) (*kustomizev1.Kustomization, error) {
	if tmpl == nil {
		return nil, errors.New("application template is empty ")
	}

	if len(params) == 0 {
		return tmpl, nil
	}

	b, err := json.Marshal(tmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Kustomization for template rendering: %w", err)
	}

	rendered, err := render(b, params)
	if err != nil {
		return nil, err
	}

	var updated kustomizev1.Kustomization
	err = json.Unmarshal(rendered, &updated)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rendered Kustomization template: %w", err)
	}

	return &updated, nil
}

func render(b []byte, params map[string]any) ([]byte, error) {
	t, err := template.New("kustomization").Funcs(funcMap).Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var out bytes.Buffer
	if err := t.Execute(&out, params); err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	return out.Bytes(), nil
}
