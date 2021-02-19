// TODO: Update the Attrbuition.txt file when adding/removing dependencies

module new-app-service

go 1.15

require (
	github.com/edgexfoundry/app-functions-sdk-go/v2 v2.0.0-dev.14
	github.com/edgexfoundry/go-mod-core-contracts/v2 v2.0.0-dev.23
	github.com/google/uuid v1.2.0
	github.com/stretchr/testify v1.7.0
)

replace github.com/edgexfoundry/app-functions-sdk-go/v2 => ../
