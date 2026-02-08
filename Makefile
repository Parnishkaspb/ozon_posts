build:
	docker compose build  && docker-compose up -d


protoc_generate:
	protoc \
	  -I proto \
	  --go_out=proto/gen \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=proto/gen \
	  --go-grpc_opt=paths=source_relative \
	  proto/service/v1/service.proto