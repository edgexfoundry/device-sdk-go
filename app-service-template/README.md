# Application Service Template

This folder contains a buildable/runnable template for a new custom application service based on the Pre-Release 2.x release of the App Functions SDK. 

> **Note**: If you only need to use the built-in pipeline functions, then it is advisable that you use `App Service Configurable` rather then create a new custom application service. See [here](https://docs.edgexfoundry.org/1.3/microservices/application/AppServiceConfigurable/) for more details on `App Service Configurable`

Follow the instructions below to create your new customer application service:

1. Copy contents of this folder to your new folder

2. Change name `new-app-service` in go.mod to an appropriate Go Module name for your service

   - Typically this is the URL to the repository for your service

3. Do a global search and replace on `new-app-service` to replace it with the name of your service

   - Note that this name is used as the service key, so it needs to use dashes rather than spaces in the name and avoid other special characters

4. Adjust your local import statements to match the name you selected in the go.mod file

   - Only needed in `main.go` if the Go Module name changed to a URL

5.  Run unit tests to verify changes didn't break the code

   - `make test`

6. Verify you are able to build the executable

   - `make build`

7. Update the `Makefile` docker build to adjust image name appropriately 

8. Verify the docker image still builds with your new image name

   - `make docker`

9. Address all the TODO's in the source files and add your custom code

10. Build and test your new custom application service

    