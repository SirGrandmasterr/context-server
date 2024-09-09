package entities

type LlmRequest struct {
	Stream           bool          `json:"stream"`
	NPredict         int           `json:"n_predict"`
	Temperature      float64       `json:"temperature"`
	Stop             []string      `json:"stop"`
	RepeatLastN      int           `json:"repeat_last_n"`
	RepeatPenalty    float64       `json:"repeat_penalty"`
	TopK             int           `json:"top_k"`
	TopP             float64       `json:"top_p"`
	MinP             float64       `json:"min_p"`
	TfsZ             int           `json:"tfs_z"`
	TypicalP         int           `json:"typical_p"`
	PresencePenalty  int           `json:"presence_penalty"`
	FrequencyPenalty int           `json:"frequency_penalty"`
	Mirostat         int           `json:"mirostat"`
	MirostatTau      int           `json:"mirostat_tau"`
	MirostatEta      float64       `json:"mirostat_eta"`
	Grammar          string        `json:"grammar"`
	NProbs           int           `json:"n_probs"`
	MinKeep          int           `json:"min_keep"`
	ImageData        []interface{} `json:"image_data"`
	CachePrompt      bool          `json:"cache_prompt"`
	APIKey           string        `json:"api_key"`
	Prompt           string        `json:"prompt"`
}

type StreamResponse struct {
	Content    string `json:"content"`
	Stop       bool   `json:"stop"`
	IDSlot     int    `json:"id_slot"`
	Multimodal bool   `json:"multimodal"`
	Index      int    `json:"index"`
}
