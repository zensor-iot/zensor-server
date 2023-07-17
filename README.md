# Server

## Dependencies

* [Emitter Go](https://github.com/emitter-io/go)
* [Goka](github.com/lovoo/goka)

## Run

To run the server you must execute next command:

`plz run //server`

## Todo

* [x] Database migrations MVP
* [ ] Database migrations templates
* [x] Materialize as query persistence layer
* [x] Device registered events
* [x] HTTP server
* [x] HTTP endpoint healthz
* [x] HTTP endpoint metrics
* [ ] HTTP resource devices
* [x] HTTP resource events
* [ ] HTTP resource sensor
* [ ] Open API documentation
* [x] Publish sensor events from mqtt to event emitted kafka topic
* [x] Add device id as key for kafka messages
* [ ] Viper Config
