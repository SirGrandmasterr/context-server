compose = docker compose

make-migrate-up:
	docker compose exec llamawhisperer bash -c './llm-whisperer migrate'

make-up:
	$(compose) up -d

make-down:
	$(compose) down

make-setup:
	$(compose) up --build -d
	make-migrate-up


#./llama-server.exe -m '.\models\tinyllama-1.1b-chat-v1.0.Q8_0(1).gguf' -c 512 -ngl 50 