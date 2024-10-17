download_big:
ifeq (,$(wildcard ./models/Llama-3.1-8B-Ultra-Instruct-Q6_K.gguf)) \
    huggingface-cli download bartowski/Llama-3.1-8B-Ultra-Instruct-GGUF --include "Llama-3.1-8B-Ultra-Instruct-Q6_K.gguf" --local-dir ./models \
endif
	

download_small:
ifeq (,$(wildcard ./models/Llama-3.2-1B-Instruct-Q8_0.gguf))
	huggingface-cli download bartowski/Llama-3.2-1B-Instruct-GGUF --include "Llama-3.2-1B-Instruct-Q8_0.gguf" --local-dir ./models
endif

download: download_small download_big