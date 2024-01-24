module dns-dso

go 1.18

replace (
	dns-client => ./dns-client
	dns-client-cache => ./dns-client-cache
	dns-connect => ./dns-connect
	dns-model => ./dns-model
	dns-process => ./dns-process
	dns-server => ./dns-server
	dns-sys => ./dns-sys
	dns-zonefile => ./dns-zonefile
	dso-core => ./dso-core
	dso-push => ./dso-push
	query-core => ./query-core
	update-core => ./update-core
)

require (
	dns-client v0.0.0-20240102083618-3f9595db54d3
	dns-server v0.0.0-20240102083910-f1f5a0576818
	dns-sys v0.0.0-20240102075912-6a8eaba60342
	dns-zonefile v0.0.0-20240102075917-115d9401bb9b
	dso-push v0.0.0-20240102075933-97fcadc84236
	github.com/cpusoft/goutil v1.0.33-0.20240114124856-3ec6d1368498
	github.com/gin-gonic/gin v1.9.1
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4
)

require (
	dns-client-cache v0.0.0-20240102080021-f3013c6caefd // indirect
	dns-connect v0.0.0-20240102075920-5665717b2d5e // indirect
	dns-model v0.0.0-20240102075808-7c5cdd7b7936 // indirect
	dns-process v0.0.0-20240102075955-9857fe87dafa // indirect
	dso-core v0.0.0-20240102075930-61d0e4988614 // indirect
	github.com/bwesterb/go-zonefile v1.0.0 // indirect
	github.com/bytedance/sonic v1.10.2 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20230717121745-296ad89f973d // indirect
	github.com/chenzhuoyu/iasm v0.9.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.16.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/guregu/null v4.0.0+incompatible // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/klauspost/cpuid/v2 v2.2.6 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-sqlite3 v1.14.19 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/parnurzeal/gorequest v0.2.17-0.20200918112808-3a0cb377f571 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shiena/ansicolor v0.0.0-20230509054315-a9deabde6e02 // indirect
	github.com/smarty/assertions v1.15.1 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
	query-core v0.0.0-20240102075925-a5a2c4d90c46 // indirect
	update-core v0.0.0-20240102075937-157616f71cce // indirect
	xorm.io/builder v0.3.13 // indirect
	xorm.io/xorm v1.3.6 // indirect
)
