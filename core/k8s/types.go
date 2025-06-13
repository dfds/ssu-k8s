package k8s

import (
	"errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"

	//"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/rules"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	Metadata v1.ObjectMeta
	Type     v1.TypeMeta
	Object   interface{}
}

type IngressRoute struct {
	v1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	v1.ObjectMeta `json:"metadata,omitempty"`

	Spec IngressRouteSpec `json:"spec,omitempty"`
}

type IngressRouteSpec struct {
	Routes []IngressRouteSpecRoute `json:"routes"`
}

type IngressRouteSpecRoute struct {
	Kind     string                         `json:"kind"`
	Match    string                         `json:"match"`
	Services []IngressRouteSpecRouteService `json:"services"`
}

func (ing *IngressRoute) PopulateDefaultsIfEmpty() {
	var routes []IngressRouteSpecRoute
	for _, route := range ing.Spec.Routes {
		var svcs []IngressRouteSpecRouteService
		for _, svc := range route.Services {
			if svc.Namespace == "" {
				svc.Namespace = ing.GetNamespace()
			}
			svcs = append(svcs, svc)
		}
		route.Services = svcs
		routes = append(routes, route)
	}
	ing.Spec.Routes = routes
}

type ExtractedRule struct {
	Host       string
	PathPrefix string
}

func ExtractHostAndPath(tree *rules.Tree) (*ExtractedRule, error) {
	info := &ExtractedRule{}

	var walker func(*rules.Tree, *ExtractedRule)
	walker = func(t *rules.Tree, inf *ExtractedRule) {
		if t == nil {
			return
		}

		if t.Matcher != "" && t.Matcher != "and" && t.Matcher != "or" {
			if (t.Matcher == "Host" || t.Matcher == "HostRegexp") && inf.Host == "" && len(t.Value) > 0 {
				inf.Host = t.Value[0]
			}

			if (t.Matcher == "PathPrefix" || t.Matcher == "Path") && len(t.Value) > 0 {
				inf.PathPrefix = t.Value[0]
			}
		}

		walker(t.RuleLeft, inf)
		walker(t.RuleRight, inf)
	}

	walker(tree, info)

	if info.Host == "" {
		return nil, errors.New("no 'Host' or 'HostRegexp' rule found in the provided tree")
	}

	return info, nil
}

func (ing *IngressRouteSpecRoute) ParseMatch() (*ExtractedRule, error) {
	parser, err := rules.NewParser([]string{"Host", "HostRegexp", "PathPrefix", "Path"})
	if err != nil {
		return nil, err
	}

	parsed, err := parser.Parse(ing.Match)
	if err != nil {
		return nil, err
	}

	treeBuilder := parsed.(rules.TreeBuilder)
	tree := treeBuilder()

	if tree == nil {
		return nil, errors.New("unable to generate tree from IngressRoute rule")
	}

	//fmt.Println(tree)

	extractedRule, err := ExtractHostAndPath(tree)
	if err != nil {
		return nil, err
	}
	return extractedRule, nil
}

type IngressRouteSpecRouteService struct {
	Kind      string             `json:"kind"`
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Port      intstr.IntOrString `json:"port"`
}

func (ing *IngressRouteSpecRouteService) GetPort() string {
	if ing.Port.Type == 1 {
		return ing.Port.StrVal
	} else {
		return strconv.Itoa(int(ing.Port.IntVal))
	}
}
