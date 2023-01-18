# Audit Log Service

This service is responsible for recording and querying logs that are sent by other systems. The service provides an Advanced Message Queuing Protocol for submitting logs data and a HTTP endpoint for querying recorded log data by field values.

## Event Types
The following are examples of logs that can be recorded:

- A new customer account was created for a given identity
- A customer performed an action on a resource
- A customer was billed a certain amount
- A customer account was deactivated

The list of log types is open-ended and new log types can be added without modifying the code. All logs should contain a common set of fields and a set of fields specific to the log type.

## Data Model
The service models an audit trail of logs with a schema that captures the invariant data content along with the variant, log-specific content. The service uses data storage mechanisms that are appropriate for the write-intensive nature of the service.

```
type Log struct {
	ID        primitive.ObjectID `bson:"_id"`
	Timestamp time.Time          `json:"created_at"`
	Action    string             `json:"action"`
	Actor     Actor              `json:"actor"`
	Entity    Entity             `json:"entity"`
	Context   Context            `json:"context"`
	Extension map[string]any     `json:"extension,omitempty"`
}
```
See [models](./internal/repository/model/model.go) for more info on the data model.

## API
The service is a microservice API that takes asychronous and synchronous approaches to solving the task of receiving, storing and retrieving logs. The API has the following HTTP endpoints:

- Service Registration
  - URL: `/v1/tokens/register`
  - Method: **POST**
  - Auth Required: No
  - Data Params: `{"service_id": "<service_identifier>"}`
  - Success Response:
    - Code: 201
    - Content: A new API key
  - Example:
    - ```curl -i -d "$DATA" http://localhost/v1/tokens/register```

- Reset Token
  - URL: `/v1/tokens/reset`
  - Method: **PATCH**
  - Auth Required: Yes
  - Data Params: `{"service_id": "<service_identifier>"}`
  - Success Response:
    - Code: 200
    - Content: A new API key
  - Example:
    - ```curl -i -d "$DATA" -H "Authorization: Key XXXX" http://localhost/v1/tokens/reset```

- Query Log 
  - URL: `/v1/logs`
  - Method: **GET**
  - Auth Required: Yes
  - Data Params: Query parameters containing the field values to filter the events by
  - Success Response:
    - Code: 200
    - Content: List of logs that match the query
  - Example:
    - ```curl -i -H "Authorization: Key XXXX" 'http://localhost/v1/logs?action=createdstart_timestamp=2022-08-16T12:34:56Z'```

- Health check
  - URL: `/v1/ping`
  - Method: **GET**
  - Auth Required: No
  - Data Params: None
  - Success Response:
    - Code: 200
    - Content: Service healthe check data
  - Example:
    - ```curl -i localhost:8081/v1/ping```

RabbitMQ is used to asynchronously handle log submission in the audit log service. This means that the service can handle a high volume of logs without being blocked by the submission process.

When a log is submitted to the service, instead of being processed immediately by the service, it is sent to a RabbitMQ queue. The service then consumes the logs from the queue at its own pace, without being overwhelmed by a large number of incoming logs.

This allows the service to handle a large number of logs in asynchronously, improving its overall performance. Additionally, it also allows the service to handle situations where the rate of incoming events exceeds the rate at which the service can process them, by temporarily storing the excess events in the queue. This ensures that the service does not become overwhelmed and can continue to function normally.

This architecture is also fault-tolerant and robust, as the queue acts as a buffer, ensuring that events are not lost even if the service is temporarily unavailable or unable to process them. See [example](./cmd/example/publisher.go) for implementation. See [run/example](#runexample) for usage.

## Deployment

The makefile included in this project provides several helpful commands to simplify the development and testing process.

The makefile includes the following commands:

### **help**

Prints a help message that lists all the commands in the makefile and their usage.

```
make help
```
### **run/service**
Runs the audit log service. It uses the `docker-compose` command to build and start the service, with the configuration specified in the `docker-compose.yaml` file.

```
make run/service
```

### **run/example**
Runs the example application. It uses the `go run` command to start the example application, which can be used to submit log to the audit log service.

```
make run/example
```
> Note; Run the `make run/service` command to setup rabbitMQ before executing `make run/example` to publish logs.

### **test suite**
Runs the test suite for the service. It uses the `go test` command to run all the tests in the project, with the -v option to display verbose output.

```
make test
```
The makefile also includes the `.envrc` file which contains variable that are used in the makefile.


## Future Work
Add scalability and operational concerns that need to be addressed in the future.

- Added monitoring and logging capabilities to the service to aid in troubleshooting and debugging.
- Implement a stronger mechanism for handling and retrying failed event submissions.
- Implement a mechanism for archiving old events to keep the data storage size manageable.
- Explore the use of a more robust data storage solution to handle the write-intensive nature of the service and improve performance.
- Consider implementing a better approach to handle and process large volumes of events in real-time.

Click [here](https://docs.google.com/document/d/1lxItFNptU2uRCxcCTuMxFWg_LDO_qJyAJ1RgiSOiaJM/edit?usp=sharing) for more.
