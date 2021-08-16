# salesforce-integration API

This document describes how to use the salesforce-integration.

### Authenticate

This endpoint allows you to obtain a token required to be able to use the resources. Requires usernames and passwords inserted in salesforce-integration env vars. 

There are following types of token in salesforce-integration, each one has a role:

*Yalo:* Description of the role
*SalesForce:* Description of the role

All resources can be accessed through this token with in an Authorization header or a queryParam. 

`POST /authenticate`

#### Request body

```json
    {
      "username" : "user",
      "password" : "password"
    }
```

| Field | Type | Required |
| :--- | :--- | :--- |
| username | `string` | Y |
| password | `string` | Y |

#### Response body

```json
    {
      "token": "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6ImEyZWY5N2RjNWQ3MWYzNjUwM2NjY2UzYTE3NjRmYmQyYzAyNjgxNDQiLCJyb2xlIjoiQURNSU5fUk9MRSIsInVzZXJuYW1lIjoiYWRtaW5Vc2VyIn0.2-CDkQwBwiAjsC5APT356Xn5TeZwSEG6LHL7Da8hBSK--qJVzMcKe4pUTbCNtlr8"  
    }
```

#### Failed response body

```json
    {
      "message": "Invalid credentials."
    }
```

### Token check

This endpoint only serves us to check that the token we want to use is valid.  

`GET /tokens/check`

#### Required role 

*Role*

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Response body

```json
    {
      "username": "user",
      "role" : "ROLE"
    }
```

#### Failed response body

```json
    {
      "message": "Invalid Authorization token."
    }
```

### Status Codes

salesforce-integration returns the following status codes:

| Status Code | Name | Description |
| :--- | :--- | :--- |
| 200 | `OK` | Request accepted and executed successfully.
| 400 | `BAD REQUEST` | Request body invalid or required parameter missing.
| 401 | `UNAUTHORIZED` | The token that is used is not valid.
| 403 | `FORBIDEN` | The token that is used does not have permission to use the resource. 
| 404 | `NOT FOUND` | Entity not found.
| 500 | `INTERNAL SERVER ERROR` | Something bad happen.

  