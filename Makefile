compose = docker compose

make migrate-up:
	compose exec llamawhisperer bash -c './llm-whisperer migrate'

make up:
	$(compose) up -d

make down:
	$(compose) down

make setup:
	$(compose) up --build -d
	make migrate-up
