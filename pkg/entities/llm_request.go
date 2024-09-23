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

type AssistantResponse struct {
	Content            string `json:"content"`
	IDSlot             int    `json:"id_slot"`
	Stop               bool   `json:"stop"`
	Model              string `json:"model"`
	TokensPredicted    int    `json:"tokens_predicted"`
	TokensEvaluated    int    `json:"tokens_evaluated"`
	GenerationSettings struct {
		NCtx             int           `json:"n_ctx"`
		NPredict         int           `json:"n_predict"`
		Model            string        `json:"model"`
		Seed             int           `json:"seed"`
		Temperature      float64       `json:"temperature"`
		DynatempRange    float64       `json:"dynatemp_range"`
		DynatempExponent float64       `json:"dynatemp_exponent"`
		TopK             int           `json:"top_k"`
		TopP             float64       `json:"top_p"`
		MinP             float64       `json:"min_p"`
		TfsZ             float64       `json:"tfs_z"`
		TypicalP         float64       `json:"typical_p"`
		RepeatLastN      int           `json:"repeat_last_n"`
		RepeatPenalty    float64       `json:"repeat_penalty"`
		PresencePenalty  float64       `json:"presence_penalty"`
		FrequencyPenalty float64       `json:"frequency_penalty"`
		Mirostat         int           `json:"mirostat"`
		MirostatTau      float64       `json:"mirostat_tau"`
		MirostatEta      float64       `json:"mirostat_eta"`
		PenalizeNl       bool          `json:"penalize_nl"`
		Stop             []interface{} `json:"stop"`
		MaxTokens        int           `json:"max_tokens"`
		NKeep            int           `json:"n_keep"`
		NDiscard         int           `json:"n_discard"`
		IgnoreEos        bool          `json:"ignore_eos"`
		Stream           bool          `json:"stream"`
		NProbs           int           `json:"n_probs"`
		MinKeep          int           `json:"min_keep"`
		Grammar          string        `json:"grammar"`
		Samplers         []string      `json:"samplers"`
	} `json:"generation_settings"`
	Prompt       string `json:"prompt"`
	Truncated    bool   `json:"truncated"`
	StoppedEos   bool   `json:"stopped_eos"`
	StoppedWord  bool   `json:"stopped_word"`
	StoppedLimit bool   `json:"stopped_limit"`
	StoppingWord string `json:"stopping_word"`
	TokensCached int    `json:"tokens_cached"`
	Timings      struct {
		PromptN             int     `json:"prompt_n"`
		PromptMs            float64 `json:"prompt_ms"`
		PromptPerTokenMs    float64 `json:"prompt_per_token_ms"`
		PromptPerSecond     float64 `json:"prompt_per_second"`
		PredictedN          int     `json:"predicted_n"`
		PredictedMs         float64 `json:"predicted_ms"`
		PredictedPerTokenMs float64 `json:"predicted_per_token_ms"`
		PredictedPerSecond  float64 `json:"predicted_per_second"`
	} `json:"timings"`
	Index int `json:"index"`
}
