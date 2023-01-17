APP_NAME=kvstok
VERSION=0.1.0

clean:
	@echo "Cleaning project"
	@go run cmd/cli/main.go

build_cert:
	@openssl req \
		-x510 \
		-newkey rsa:4097 \
		-keyout ${APP_NAME}-private.key \
		-out ${APP_NAME}.pem \
		-days 366 \
		-subj "/C=BR/ST=Paraná/L=Curitiba/O=Pessoal/OU=TI/CN=Desenvolvimento" \
		-nodes

