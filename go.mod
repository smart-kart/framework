module github.com/smart-kart/framework

go 1.24.4

require (
	github.com/go-playground/validator/v10 v10.24.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
	github.com/jackc/pgx/v5 v5.7.6
	github.com/smart-kart/proto v0.0.0
	github.com/spf13/cast v1.10.0
	github.com/stretchr/testify v1.8.4
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250929231259-57b25ae835d4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250922171735-9219d122eba9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/smart-kart/proto => ../proto
