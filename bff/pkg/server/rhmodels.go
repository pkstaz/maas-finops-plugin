package server

import "strings"

type CatalogModel struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	Family      string  `json:"family"`
	ParamsB     float64 `json:"paramsB"`
	Quant       string  `json:"quant"`
	URI         string  `json:"uri"`
	Source      string  `json:"source"`
}

// Red Hat AI models commonly served on RHOAI (Hugging Face RedHatAI + Granite).
var redHatModels = []CatalogModel{
	rh("granite-3.3-8b-instruct", "Granite 3.3 8B Instruct", "Granite", 8, "w4a16", "hf://ibm-granite/granite-3.3-8b-instruct"),
	rh("granite-3.2-8b-instruct", "Granite 3.2 8B Instruct", "Granite", 8, "w4a16", "hf://ibm-granite/granite-3.2-8b-instruct"),
	rh("granite-3.1-8b-instruct", "Granite 3.1 8B Instruct", "Granite", 8, "w4a16", "hf://ibm-granite/granite-3.1-8b-instruct"),
	rh("granite-8b-code-instruct", "Granite 8B Code Instruct", "Granite", 8, "w4a16", "hf://ibm-granite/granite-8b-code-instruct"),
	rh("llama-3.2-1b-instruct", "Llama 3.2 1B Instruct", "Llama", 1, "w4a16", "hf://RedHatAI/Llama-3.2-1B-Instruct-quantized.w4a16"),
	rh("llama-3.2-3b-instruct", "Llama 3.2 3B Instruct", "Llama", 3, "w4a16", "hf://RedHatAI/Llama-3.2-3B-Instruct-quantized.w4a16"),
	rh("llama-3.1-8b-instruct-w4a16", "Llama 3.1 8B Instruct (W4A16)", "Llama", 8, "w4a16", "hf://RedHatAI/Meta-Llama-3.1-8B-Instruct-quantized.w4a16"),
	rh("llama-3.1-8b-instruct-fp8", "Llama 3.1 8B Instruct (FP8)", "Llama", 8, "fp8", "hf://RedHatAI/Meta-Llama-3.1-8B-Instruct-FP8"),
	rh("llama-3.1-70b-instruct-w4a16", "Llama 3.1 70B Instruct (W4A16)", "Llama", 70, "w4a16", "hf://RedHatAI/Meta-Llama-3.1-70B-Instruct-quantized.w4a16"),
	rh("llama-3.3-70b-instruct-w4a16", "Llama 3.3 70B Instruct (W4A16)", "Llama", 70, "w4a16", "hf://RedHatAI/Llama-3.3-70B-Instruct-quantized.w4a16"),
	rh("mistral-7b-instruct-w4a16", "Mistral 7B Instruct (W4A16)", "Mistral", 7, "w4a16", "hf://RedHatAI/Mistral-7B-Instruct-v0.3-quantized.w4a16"),
	rh("mistral-small-24b-w4a16", "Mistral Small 24B (W4A16)", "Mistral", 24, "w4a16", "hf://RedHatAI/Mistral-Small-24B-Instruct-2501-quantized.w4a16"),
	rh("mixtral-8x7b-instruct-w4a16", "Mixtral 8x7B Instruct (W4A16)", "Mixtral", 47, "w4a16", "hf://RedHatAI/Mixtral-8x7B-Instruct-v0.1-quantized.w4a16"),
	rh("qwen2.5-7b-instruct-w4a16", "Qwen 2.5 7B Instruct (W4A16)", "Qwen", 7, "w4a16", "hf://RedHatAI/Qwen2.5-7B-Instruct-quantized.w4a16"),
	rh("qwen2.5-14b-instruct-w4a16", "Qwen 2.5 14B Instruct (W4A16)", "Qwen", 14, "w4a16", "hf://RedHatAI/Qwen2.5-14B-Instruct-quantized.w4a16"),
	rh("qwen2.5-32b-instruct-w4a16", "Qwen 2.5 32B Instruct (W4A16)", "Qwen", 32, "w4a16", "hf://RedHatAI/Qwen2.5-32B-Instruct-quantized.w4a16"),
	rh("qwen3-8b-w4a16", "Qwen 3 8B (W4A16)", "Qwen", 8, "w4a16", "hf://RedHatAI/Qwen3-8B-quantized.w4a16"),
	rh("phi-4-w4a16", "Phi-4 (W4A16)", "Phi", 14, "w4a16", "hf://RedHatAI/phi-4-quantized.w4a16"),
	rh("gemma-2-9b-it-w4a16", "Gemma 2 9B IT (W4A16)", "Gemma", 9, "w4a16", "hf://RedHatAI/gemma-2-9b-it-quantized.w4a16"),
	rh("gemma-2-27b-it-w4a16", "Gemma 2 27B IT (W4A16)", "Gemma", 27, "w4a16", "hf://RedHatAI/gemma-2-27b-it-quantized.w4a16"),
}

func rh(id, display, family string, paramsB float64, quant, uri string) CatalogModel {
	return CatalogModel{
		ID: id, DisplayName: display, Family: family,
		ParamsB: paramsB, Quant: quant, URI: uri, Source: "redhat-ai",
	}
}

func findCatalogModel(id string) (CatalogModel, bool) {
	for _, m := range redHatModels {
		if m.ID == id || strings.EqualFold(m.DisplayName, id) {
			return m, true
		}
	}
	return CatalogModel{}, false
}
