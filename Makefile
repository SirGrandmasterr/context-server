compose = docker compose

make-migrate-up:
	$(compose) exec llamawhisperer sh -c './llm-whisperer migrate'

make-up:
	$(compose) up -d

make-down:
	$(compose) down

make-setup:
	$(compose) up --build -d
	$(MAKE) make-migrate-up

make-delete-db:
	$(compose) exec -it mongodb mongosh --authenticationDatabase admin -u user -p pass --eval 'db.getSiblingDB("llamadrama").dropDatabase()'


#./llama-server.exe -m '.\models\tinyllama-1.1b-chat-v1.0.Q8_0(1).gguf' -c 512 -ngl 50 
#./llama-server.exe -m .\models\Meta-Llama-3.1-8B-Instruct-Q6_K.gguf -c 3500 -ngl 50
#./llama-server.exe -m .\models\Llama-3.2-1B-Instruct-Q8_0.gguf -c 2048 -ngl 50