package server

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (c *clusterClient) inventory(ctx context.Context) ClusterInventory {
	inv := ClusterInventory{Provider: "unknown", Source: "cluster", Nodes: []NodeInventory{}, Models: []ServingModel{}}
	if c == nil {
		inv.Message = "no cluster client"
		return inv
	}

	if infra, err := c.dynamic.Resource(gvrInfra).Get(ctx, "cluster", metav1.GetOptions{}); err == nil {
		inv.Platform, _, _ = unstructured.NestedString(infra.Object, "status", "platform")
		if inv.Platform == "" {
			inv.Platform, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "type")
		}
		inv.Cloud, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "azure", "cloudName")
		inv.Region, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "azure", "region")
		if inv.Region == "" {
			inv.Region, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "aws", "region")
		}
		if inv.Region == "" {
			inv.Region, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "ibmcloud", "location")
		}
		if inv.Cloud == "" {
			inv.Cloud, _, _ = unstructured.NestedString(infra.Object, "status", "platformStatus", "ibmcloud", "providerType")
		}
	}

	nodes, err := c.typed.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		for i := range nodes.Items {
			inv.Nodes = append(inv.Nodes, nodeFromK8s(&nodes.Items[i]))
		}
	}
	if inv.Region == "" {
		for _, n := range inv.Nodes {
			if n.Region != "" {
				inv.Region = n.Region
				break
			}
		}
	}

	inv.Provider = normalizeProvider(inv.Platform)
	if inv.Provider == "" {
		inv.Provider = "unknown"
	}
	if providerSupported(inv.Provider) {
		inv.Supported = true
	} else {
		inv.Supported = false
		if inv.Platform == "" {
			inv.Message = "could not read Infrastructure; enter machine cost manually"
		} else {
			inv.Message = "cloud " + inv.Platform + " has no automatic list price; use manual cost"
		}
	}

	inv.Models = c.listServingModels(ctx, inv.Nodes)
	return inv
}

func nodeFromK8s(n *corev1.Node) NodeInventory {
	labels := n.Labels
	out := NodeInventory{
		Name: n.Name,
		InstanceType: firstNonEmpty(
			labels["node.kubernetes.io/instance-type"],
			labels["beta.kubernetes.io/instance-type"],
			labels["ibm-cloud.kubernetes.io/machine-type"],
			labels["ibm-cloud.kubernetes.io/flavor"],
		),
		Region:     firstNonEmpty(labels["topology.kubernetes.io/region"], labels["failure-domain.beta.kubernetes.io/region"]),
		Zone:       firstNonEmpty(labels["topology.kubernetes.io/zone"], labels["failure-domain.beta.kubernetes.io/zone"]),
		GPUProduct: strings.ReplaceAll(labels["nvidia.com/gpu.product"], "_", " "),
		GPUFamily:  labels["nvidia.com/gpu.family"],
	}
	if _, ok := labels["node-role.kubernetes.io/gpu"]; ok {
		out.Role = "gpu"
	} else if _, ok := labels["node-role.kubernetes.io/master"]; ok {
		out.Role = "master"
	} else if _, ok := labels["node-role.kubernetes.io/control-plane"]; ok {
		out.Role = "master"
	} else {
		out.Role = "worker"
	}
	out.GPUCount, _ = strconv.Atoi(labels["nvidia.com/gpu.count"])
	out.GPUMemoryMiB, _ = strconv.Atoi(labels["nvidia.com/gpu.memory"])
	if q, ok := n.Status.Allocatable["nvidia.com/gpu"]; ok && out.GPUCount == 0 {
		out.GPUCount = int(q.Value())
	}
	if out.GPUCount > 0 && out.Role == "worker" {
		out.Role = "gpu"
	}
	return out
}

func (c *clusterClient) listServingModels(ctx context.Context, nodes []NodeInventory) []ServingModel {
	if c == nil {
		return nil
	}
	discovered, err := c.listModels(ctx)
	if err != nil {
		discovered = nil
	}
	llms := map[string]unstructured.Unstructured{}
	if list, err := c.dynamic.Resource(gvrLLM).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{}); err == nil {
		for i := range list.Items {
			item := list.Items[i]
			llms[item.GetNamespace()+"/"+item.GetName()] = item
			if modelName, _, _ := unstructured.NestedString(item.Object, "spec", "model", "name"); modelName != "" {
				llms[item.GetNamespace()+"/"+modelName] = item
			}
		}
	}
	firstGPU := ""
	for _, n := range nodes {
		if n.GPUCount > 0 && firstGPU == "" {
			firstGPU = n.Name
		}
	}
	out := make([]ServingModel, 0, len(discovered))
	for _, d := range discovered {
		m := ServingModel{
			Name: d.Name, Namespace: d.Namespace, DisplayName: d.DisplayName,
			Kind: d.Kind, Origin: firstNonEmpty(d.Origin, originFromKind(d.Kind)),
			Provider: d.Provider, Endpoint: d.Endpoint, TargetModel: d.TargetModel,
		}
		if isExternalOrigin(m.Origin, m.Kind) {
			out = append(out, m)
			continue
		}
		item, ok := llms[d.Namespace+"/"+d.Name]
		if !ok {
			out = append(out, m)
			continue
		}
		m.URI, _, _ = unstructured.NestedString(item.Object, "spec", "model", "uri")
		if m.DisplayName == "" {
			m.DisplayName = m.Name
		}
		m.Quant = detectQuant(m.URI + " " + m.DisplayName + " " + m.Name)
		m.ParamsB = detectParamsB(m.URI + " " + m.DisplayName + " " + m.Name)
		containers, _, _ := unstructured.NestedSlice(item.Object, "spec", "template", "containers")
		for _, raw := range containers {
			cm, _ := raw.(map[string]interface{})
			m.GPURequest += gpuQuantity(cm)
			if args, ok := cm["args"].([]interface{}); ok {
				m.MaxModelLen = argInt(args, "--max-model-len")
			}
		}
		m.NodeName = c.podNodeForModel(ctx, m.Namespace, item.GetName(), firstGPU)
		out = append(out, m)
	}
	return out
}

func (c *clusterClient) podNodeForModel(ctx context.Context, ns, name, fallback string) string {
	if c == nil || ns == "" {
		return fallback
	}
	pods, err := c.typed.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fallback
	}
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Spec.NodeName == "" {
			continue
		}
		if strings.Contains(p.Name, name) {
			for _, ctn := range p.Spec.Containers {
				if _, ok := ctn.Resources.Limits["nvidia.com/gpu"]; ok {
					return p.Spec.NodeName
				}
			}
			return p.Spec.NodeName
		}
	}
	return fallback
}

func gpuQuantity(container map[string]interface{}) float64 {
	for _, path := range [][]string{
		{"resources", "limits", "nvidia.com/gpu"},
		{"resources", "requests", "nvidia.com/gpu"},
	} {
		if v, ok, _ := unstructured.NestedFieldCopy(container, path...); ok {
			switch t := v.(type) {
			case int64:
				return float64(t)
			case float64:
				return t
			case string:
				f, _ := strconv.ParseFloat(t, 64)
				return f
			}
		}
	}
	return 0
}

func argInt(args []interface{}, flag string) int {
	for i, a := range args {
		s, _ := a.(string)
		if s == flag && i+1 < len(args) {
			n, _ := strconv.Atoi(fmtString(args[i+1]))
			return n
		}
		if strings.HasPrefix(s, flag+"=") {
			n, _ := strconv.Atoi(strings.TrimPrefix(s, flag+"="))
			return n
		}
	}
	return 0
}

func fmtString(v interface{}) string {
	s, _ := v.(string)
	return s
}

var paramsRE = regexp.MustCompile(`(?i)(?:^|[^\d])(\d+(?:\.\d+)?)[ ]*[bB](?:\b|-)`)

func detectParamsB(s string) float64 {
	low := strings.ToLower(s)
	if strings.Contains(low, "llama-3.1-8") || strings.Contains(low, "llama-31") || strings.Contains(low, "llama3.1-8") {
		return 8
	}
	m := paramsRE.FindStringSubmatch(s)
	if len(m) == 2 {
		f, _ := strconv.ParseFloat(m[1], 64)
		return f
	}
	return 0
}

func detectQuant(s string) string {
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "w4a16") || strings.Contains(low, "awq") || strings.Contains(low, "gptq") || strings.Contains(low, "int4"):
		return "w4a16"
	case strings.Contains(low, "w8a8") || strings.Contains(low, "int8"):
		return "w8a8"
	case strings.Contains(low, "fp8"):
		return "fp8"
	case strings.Contains(low, "bf16"):
		return "bf16"
	case strings.Contains(low, "fp16"):
		return "fp16"
	default:
		return "unknown"
	}
}
