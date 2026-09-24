package server

import (
	"context"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	gvrModelRef = schema.GroupVersionResource{
		Group: "maas.opendatahub.io", Version: "v1alpha1", Resource: "maasmodelrefs",
	}
	gvrSubscription = schema.GroupVersionResource{
		Group: "maas.opendatahub.io", Version: "v1alpha1", Resource: "maassubscriptions",
	}
	gvrLLM = schema.GroupVersionResource{
		Group: "serving.kserve.io", Version: "v1alpha1", Resource: "llminferenceservices",
	}
	gvrMaaSExternal = schema.GroupVersionResource{
		Group: "maas.opendatahub.io", Version: "v1alpha1", Resource: "externalmodels",
	}
	gvrInfra = schema.GroupVersionResource{
		Group: "config.openshift.io", Version: "v1", Resource: "infrastructures",
	}
)

type clusterClient struct {
	typed   kubernetes.Interface
	dynamic dynamic.Interface
	ns      string
}

type discoveredModel struct {
	Name        string
	Namespace   string
	DisplayName string
	Phase       string
	Kind        string
	Origin      string
	Provider    string
	Endpoint    string
	TargetModel string
}

type discoveredSub struct {
	Name       string
	Namespace  string
	Priority   int64
	Models     []string
	RateLimits []string
}

func newClusterClient() (*clusterClient, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		cfg, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	typed, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	ns := os.Getenv("POD_NAMESPACE")
	if ns == "" {
		ns = "maas-finops-plugin"
	}
	return &clusterClient{typed: typed, dynamic: dyn, ns: ns}, nil
}

func (c *clusterClient) listModels(ctx context.Context) ([]discoveredModel, error) {
	if c == nil {
		return nil, fmt.Errorf("no cluster client")
	}
	seen := map[string]discoveredModel{}
	externals := c.indexExternalModels(ctx)

	refs, err := c.dynamic.Resource(gvrModelRef).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range refs.Items {
			name := item.GetName()
			ns := item.GetNamespace()
			phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")
			refKind, _, _ := unstructured.NestedString(item.Object, "spec", "modelRef", "kind")
			refName, _, _ := unstructured.NestedString(item.Object, "spec", "modelRef", "name")
			if refName == "" {
				refName = name
			}
			display := item.GetAnnotations()["openshift.io/display-name"]
			if display == "" {
				display = name
			}
			kind := firstNonEmpty(refKind, "MaaSModelRef")
			m := discoveredModel{
				Name: name, Namespace: ns, DisplayName: display, Phase: phase,
				Kind: kind, Origin: originFromKind(kind), TargetModel: name,
			}
			if info, ok := externals[ns+"/"+refName]; ok {
				m.Provider = info.Provider
				m.Endpoint = info.Endpoint
				m.TargetModel = firstNonEmpty(info.TargetModel, name)
				if display == name && info.DisplayName != "" {
					m.DisplayName = info.DisplayName
				}
			}
			if m.Origin == originExternal && m.Provider == "" {
				m.Provider = "external"
			}
			seen[ns+"/"+name] = m
		}
	}

	llms, err := c.dynamic.Resource(gvrLLM).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, item := range llms.Items {
			name := item.GetName()
			ns := item.GetNamespace()
			key := ns + "/" + name
			display := item.GetAnnotations()["openshift.io/display-name"]
			modelName, _, _ := unstructured.NestedString(item.Object, "spec", "model", "name")
			if modelName == "" {
				modelName = name
			}
			phase, _, _ := unstructured.NestedString(item.Object, "status", "url")
			if existing, ok := seen[key]; ok {
				if display != "" && existing.DisplayName == existing.Name {
					existing.DisplayName = display
				}
				if existing.Kind == "" || existing.Kind == "MaaSModelRef" {
					existing.Kind = "LLMInferenceService"
				}
				seen[key] = existing
				continue
			}
			if display == "" {
				display = modelName
			}
			urlPhase := "Available"
			if phase == "" {
				urlPhase = "Pending"
			}
			seen[key] = discoveredModel{
				Name: modelName, Namespace: ns, DisplayName: display, Phase: urlPhase,
				Kind: "LLMInferenceService", Origin: originRHOAI,
			}
		}
	}

	out := make([]discoveredModel, 0, len(seen))
	for _, m := range seen {
		out = append(out, m)
	}
	return out, nil
}

func (c *clusterClient) listSubscriptions(ctx context.Context) ([]discoveredSub, error) {
	if c == nil {
		return nil, fmt.Errorf("no cluster client")
	}
	list, err := c.dynamic.Resource(gvrSubscription).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	out := make([]discoveredSub, 0, len(list.Items))
	for _, item := range list.Items {
		sub := discoveredSub{Name: item.GetName(), Namespace: item.GetNamespace()}
		if p, ok, _ := unstructured.NestedInt64(item.Object, "spec", "priority"); ok {
			sub.Priority = p
		}
		refs, _, _ := unstructured.NestedSlice(item.Object, "spec", "modelRefs")
		for _, raw := range refs {
			m, _ := raw.(map[string]interface{})
			name, _ := m["name"].(string)
			if name != "" {
				sub.Models = append(sub.Models, name)
			}
			limits, _ := m["tokenRateLimits"].([]interface{})
			for _, lraw := range limits {
				lm, _ := lraw.(map[string]interface{})
				limit := fmt.Sprint(lm["limit"])
				window, _ := lm["window"].(string)
				if limit != "" && window != "" {
					sub.RateLimits = append(sub.RateLimits, limit+" / "+window)
				}
			}
		}
		out = append(out, sub)
	}
	return out, nil
}

func (c *clusterClient) getConfigMap(ctx context.Context, name string) (*corev1.ConfigMap, error) {
	if c == nil {
		return nil, fmt.Errorf("no cluster client")
	}
	return c.typed.CoreV1().ConfigMaps(c.ns).Get(ctx, name, metav1.GetOptions{})
}

func (c *clusterClient) upsertConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	if c == nil {
		return fmt.Errorf("no cluster client")
	}
	existing, err := c.typed.CoreV1().ConfigMaps(c.ns).Get(ctx, cm.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = c.typed.CoreV1().ConfigMaps(c.ns).Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}
	existing.Data = cm.Data
	_, err = c.typed.CoreV1().ConfigMaps(c.ns).Update(ctx, existing, metav1.UpdateOptions{})
	return err
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

type externalInfo struct {
	Provider    string
	Endpoint    string
	TargetModel string
	DisplayName string
}

func (c *clusterClient) indexExternalModels(ctx context.Context) map[string]externalInfo {
	out := map[string]externalInfo{}
	if c == nil {
		return out
	}
	list, err := c.dynamic.Resource(gvrMaaSExternal).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return out
	}
	for _, item := range list.Items {
		provider, _, _ := unstructured.NestedString(item.Object, "spec", "provider")
		endpoint, _, _ := unstructured.NestedString(item.Object, "spec", "endpoint")
		target, _, _ := unstructured.NestedString(item.Object, "spec", "targetModel")
		out[item.GetNamespace()+"/"+item.GetName()] = externalInfo{
			Provider:    provider,
			Endpoint:    endpoint,
			TargetModel: target,
			DisplayName: item.GetAnnotations()["openshift.io/display-name"],
		}
	}
	return out
}
