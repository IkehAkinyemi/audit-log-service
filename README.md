# Audit Log Service 📻

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
  - Data Params: [Query parameters](./internal/utils/filter.go) containing the field values to filter the log by.
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
    - ```curl -i http://localhost/v1/ping```

RabbitMQ is used to asynchronously handle log submission in the audit log service, meaning service can handle a high volume of logs without being blocked by the submission process.

This architecture is also fault-tolerant and robust, as the queue acts as a buffer, ensuring that logs are not lost even if the service is temporarily unavailable or unable to process them. See [example](./cmd/example/publisher.go) for implementation. See [run/example](#runexample) for usage.

## Prerequisites
- Go version 1.13 or higher
- [Docker](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-on-ubuntu-20-04) and [docker-compose](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-docker-compose-on-ubuntu-20-04)
- [make(1)](https://man7.org/linux/man-pages/man1/make.1.html) utility

## Deployment

The makefile included in this project provides several helpful commands to simplify the deployment and testing process.

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
The makefile also includes the `.envrc` file which contains variable that are used in the makefile

### **Query logs**
To retrieve stored logs, you will first need to obtain an API Key using the **`/v1/register`** endpoint:

```
curl -i -d '{"service_id": "platform-12345"}' http://localhost/v1/tokens/register
```
Use the returned API key within the Authorization header in the next curl command to query logs like below:

```
curl -i -H "Authorization: Key <YOUR_API_KEY>" 'http://localhost/v1/logs?action=created&start_timestamp=2022-08-16T12:34:56Z'
```

## Future Work
- Added scalability and operational concerns that need to be addressed in the future.
  - Add/extend monitoring and observability capabilities to the service to aid troubleshooting and debugging.
  - Extending RabbitMQ to add a more robust mechanism for handling and retrying failed log submissions and consumption
  - Implement a mechanism for archiving or deleting old logs to keep the data storage size manageable.
  - Explore the use of a more robust data storage solution to handle the write-intensive nature of the service and improve performance.
  - Consider implementing a better approach to handling and processing large volumes of real-time logs.
  - Removing any unnecessary functionalities from the service, like authentication and querying. This allows the service to focus more on the write-ops domain.


- More (extensive) Testing
  - create and run table-driven unit tests and sub-tests to test cover more test scenarios. 
  - unit test the HTTP handlers and middleware.
  - perform ‘end-to-end’ testing of the web service routes, middleware, and handlers.
  - create mocks of the database models and use them in unit tests.
  - use a test instance of MongoDB to perform integration tests.
  - calculate and profile code coverage for the test suite.


Click [here](https://docs.google.com/document/d/1lxItFNptU2uRCxcCTuMxFWg_LDO_qJyAJ1RgiSOiaJM/edit?usp=sharing) for more on the rationale, tradeoffs, and future works.
