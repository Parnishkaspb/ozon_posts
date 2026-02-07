module github.com/Parnishkaspb/ozon_posts_graphql

go 1.24.0

toolchain go1.24.12

require (
	github.com/99designs/gqlgen v0.17.86
	github.com/Parnishkaspb/ozon_posts_proto v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/graph-gophers/dataloader v5.0.0+incompatible
	github.com/vektah/gqlparser/v2 v2.5.31
	google.golang.org/grpc v1.78.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/sosodev/duration v1.3.1 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/Parnishkaspb/ozon_posts_proto => ../proto
