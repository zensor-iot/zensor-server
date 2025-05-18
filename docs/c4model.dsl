workspace "Zensor Server" "The core of the Zensor Ecosystem" {

    !identifiers hierarchical

    model {
        u = person "User"
        zensor_server = softwareSystem "Zensor Server" {
            wa = container "Web Application" {
                device_controller = component "Device Controller" "control_plane/httpapi" {
                    tags "controller"
                }

                device_service = component "Device Service" "control_plane/usecases" {
                    tags "usecases"
                }

                command_publisher = component "Command Publisher" "control_plane/communication" {
                    tags "communication"
                }

                command_worker = component "Command Worker" "data_plane/workers" {
                    tags "workers"
                }
            }

            msg_broker = container "Message Broker" {
                tags "PubSub"
            }

            database = container "Database" {
                tags "Database"
            }

            database -> msg_broker "read and materialize events into records"
            wa.device_controller -> wa.device_service "queue command sequence"
            wa.device_service -> wa.command_publisher "dispatch command events"
            wa.command_publisher -> msg_broker "publish pending command events"
            wa.command_worker -> database "read pending commands"
            wa.command_worker -> msg_broker "publish ready command events"
        }

        u -> zensor_server.wa.device_controller "send a sequence of commands"
        zensor_server.wa -> zensor_server.database "Reads from"


    }

    views {
        systemContext zensor_server "Diagram1" {
            include *
            autolayout lr
        }

        container zensor_server "Diagram2" {
            include *
            autolayout lr
        }

        component zensor_server.wa "Diagram3" {
            include *
            autolayout lr
        }

        styles {
            element "Element" {
                color #ffffff
            }
            element "Person" {
                background #1168bd
                shape person
            }
            element "Software System" {
                background #1168bd
            }
            element "Container" {
                background #1168bd
            }
            element "Component" {
                background #1168bd
            }
            element "Database" {
                shape cylinder
            }
            element "PubSub" {
                shape Pipe
            }
        }
    }

    configuration {
        scope softwaresystem
    }

}
