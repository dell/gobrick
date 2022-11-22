module github.com/dell/gobrick

go 1.18

require (
	github.com/dell/goiscsi v1.5.0
	github.com/dell/gonvme v1.2.1-0.20221111064610-e2ea406d3203
	github.com/golang/mock v1.3.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.8.1
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace (
	github.com/opiproject/goopicsi => ../goopicsi
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/opiproject/goopicsi v0.0.0-20221118191143-2cc8e295e23d // indirect
	github.com/opiproject/opi-api v0.0.0-20221117170559-ca2c25b808b3 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.0.0-20221002022538-bcab6841153b // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	google.golang.org/genproto v0.0.0-20220930163606-c98284e70a91 // indirect
	google.golang.org/grpc v1.51.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)
