module github.com/phanvanhai/device-service-support

go 1.12

require (
	github.com/edgexfoundry/device-sdk-go v1.2.0
	github.com/edgexfoundry/go-mod-core-contracts v0.1.54
	github.com/stianeikeland/go-rpio v4.2.0+incompatible
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

replace github.com/edgexfoundry/device-sdk-go => github.com/phanvanhai/device-sdk-go v1.2.1-0.20200428033101-76af95978ead
