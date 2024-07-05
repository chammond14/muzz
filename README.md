# Muzz Take Home Test

A RESTful web service written in Go providing a backend for a primitive dating app. It uses a PostgreSQL datastore.

The service supports creating (random) profiles, logging in, discovering other profiles and swiping on profiles.

# Contents
1. [Prerequisites](#Prerequisites)
2. [Building and Running](#Building-and-Running)
3. [Configuration](#Configuration)
4. [Using the Service](#Using-the-Service)
5. [Notes](#Notes)


# Prerequisites

Docker must be installed

[Installing Docker](https://docs.docker.com/engine/install/)

# Building and Running

This service and it's datastore are containerised with Docker, and should not require any other tools in order to run.

Navigate in the terminal to the directory where this repository has been cloned, and then run the following command:

```
docker compose up -d
```
This creates a docker container for both the datastore and the web service.

## Running Tests

The application contains automated tests. They are by no means exhaustive due to time constraints. They can be run using the following command:

```
go test -v ./...
```

**Please Note: Postgres container must be running in order to successfully run tests**

The tests use a test database to avoid corrupting application state.

# Configuration
Once running, the service is available by default at the following url
```
http://localhost:8080
```
Several variables are configurable via the .env file, included in the repo
### ENV variables
| variable        | Description           
| ------------- |:-------------:|
| ADDR      | The address and port where the service runs. Due to running via docker compose, the IP address MUST NOT be changed. The port can be altered as desired |
| DB_CONN_STRING      | Connection string for connecting to the PostgreSQL datastore. Required for functioning of the service     | 
| DB_TIMEOUT_SECONDS      | The number of seconds to allow a DB query to run for before cancelling the operation via the context. |
| isLocal      | Set to true to enable database seeding during server startup     | 


# Using the Service
### Seed Data
As mentioned above, the database will be seeded with a small set of profiles during startup. To add seed data, Postgres queries can be added to the file ```/internal/db/seed-queries.go```.

Any new queries added must be executed by adding it in the `seedDatabase` function in the `/internal/db/schema.go` file.

### Endpoints

#### `GET /user/create`
Creates a random profile in the datastore, which will be returned.

#### `POST /login`
Login as a user to the web service. The request body must take the form:

    {
        "username": "john@muzz.com",
        "password": "papayas"
    }

A successful login attempt will return a session token which must be added to the `session` header in order to make subsequent calls to `/discover` or `/swipe`

#### `POST /discover`
Discover other profiles. The request body can be used to filter the results returned. A `session` header must be attached to this request to authenticate the logged in user.

```
// header
"session": "12345678"
```

    // request body
    {
        "minAge": 40,
        "maxAge": 60,
        "genders": ["female"], // any of "male", "female", "other"
        "lat":  -0.14161508885288424, // required
        "long": 51.50149354607873 // required
    }

The lat and long values are required - this is in order to sort results in order of proximity to the user.

#### `POST /swipe`
Swipe on a profile with the given id. A `session` header must be attached to this request to authenticate the logged in user.

```
// header
"session": "12345678"
```

    // request body
    {
    "user": 4, // required
    "liked": true // required
    }

# Notes

As a general note, this task was used as an opportunity to try out PostgreSQL, and likely contains some suboptimal implementation.

### db package file structure
Logic within the db package has been split into separate files to be a little easier on the eyes. The three files with query logic are `profile.go`, `swipe.go`, and `discover.go`. This package also contains logic to apply the database schema, which would ideally be moved into the `.deployment/local` file.

### Creating Profiles

Randomly generated profiles will all have the same location. This was hardcoded for simplicity and time saving.

### Authentication

The session token supplied as a header is used to look up the userId in middleware. This prevents situations where a valid session token can be used to act on behalf of another user.


