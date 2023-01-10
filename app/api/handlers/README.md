# salesforce-integration API

This document describes how to use the different endpoints available for salesforce-integration.

The main endpoints are the creation of a chat between a Yalo bot and Salesforce and the webhooks endpoints for the Facebook and WhatsApp bots.

### Authenticate

This endpoint allows you to obtain a token required to be able to use the resources. Requires usernames and passwords inserted in salesforce-integration enviroment vars. 

There are following types of token in salesforce-integration, each one has a role:

*Yalo:* It has access to the resources for creating and ending chats, consulting the context stored in redis, among others that Yalo requires. 
*SalesForce:* This Role allows you access to some resources that Yalo allows you to access Salesforce. 

Most of the resources can be accessed through the token obtained by this resource, if it is sent in an authorization header or a queryParam and its ROLE has permission. 

**NOTE:** The only resources that cannot be accessed by this token are the webhooks registered in API integrations. 

`POST /v1/authenticate`

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
  "token": ${token}  
}
```

#### Failed response body

```json
{
  "ErrorDescription": "Invalid credentials."
}
```

### Token check

This endpoint only serves us to check that the token we want to use is valid.  

`GET /v1/tokens/check`

#### Required role 

***YALO_ROLE or SALESFORCE_ROLE***

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
  "ErrorDescription": "Invalid Authorization token."
}
```

### Health check

This resource validates that the integration service is working correctly.

`GET /v1/welcome`

#### Required role 

Not required

#### Response body 

##### 200 Status

```json
{
  "Message": "Welcome to API!"
}   
```

### Create Chat

This resource will create a chat between a Yalo bot and Salesforce. 

This endpoint will be invoked by a Yalo Studio webhook or lambda in the flow of a bot, when the user writes in the bot the option corresponding to ***Contactar a un asesor***, for example it can be the keyword **Ayuda** and create the chat with a Salesforce agent.

`POST /v1/chats/connect`

#### Required role 

***YALO_ROLE***

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Request body

##### Example chat from whatsapp bot
```json
{
  "userID" : "5215222545142",
  "botSlug" : "yalo-wa-bot",
  "botId" : "5215522114455",
  "name" : "Eduardo Ochoa",
  "provider" : "whatsapp",
  "email" : "ochoa@example.com",
  "phoneNumber" : "5215222545142",
  "extraData" : {
      "source_flow_bot" : "SFB001",
      "address": "Calle Guadalupe Victoria 13, col. Revolucion",
      "order_traking_id": "1225222565"
    }
}
```

##### Example chat from facebook bot
```json
{
  "userID" : "6199190013487244",
  "botSlug" : "yalo-msn-bot",
  "botId" : "233726951928152",
  "name" : "Eduardo Ochoa",
  "provider" : "facebook",
  "email" : "ochoa@example.com",
  "extraData" : {
      "source_flow_bot" : "SFB001",
      "address": "Calle Guadalupe Victoria 13, col. Revolucion",
      "order_traking_id": "1225222565"
    }
}
```

| Field | Type | Required | Description
| :--- | :--- | :--- | :--- |
| userID | `string` | Y | User identifier phone in whatsapp and facebook Id in messenger. |
| botSlug | `string` | Y | Bot name. |
| botId | `string` | Y | Phone for whatsapp bot or pageId in messenger bot. |
| name | `string` | Y | User name. |
| provider | `string` | Y | Chat origin, the allowed values are **whatsapp** or **facebook**. |
| email | `string` | Y | User's email to search or register the contact or person account in Salesforce.|
| phoneNumber | `string` | N | User's phone number to search or register the contact or personal account in Salesforce allowed in whatsapp bot, only if it is enabled. |
| extraData | `object` | N | It mainly sends the custom fields that the customer has in Salesforce to add them to the cases and they are sent if we have the information. If necessary you can send metadata for custom implementations. |

#### Response body 

##### 200 Status

```json
{
  "Message": "create chat successfuly"
}
```

#### Failed response body

```json
{
  "ErrorDescription": "could not create chat : Error message"
}
```

### Please check the webhook requirements for whastapp bot and for facebook Bot at [Salesforce-Integrations-Endpoints](/docs/Salesforce-Integrations-Endpoints.md) documentation. ###


### End Chat
This endpoint is on charge of finishing the chat according with the usedID associated, only if a chat exists.

`GET /v1/chat/finish/{{user_id}}`

#### Required role 

***YALO_ROLE***

#### Path params

| Name | Value | Required |
| :--- | :--- | :--- |
| user_id | `521125486585` or `facebookId` | Y |

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Response body 

##### 200 Status

```json
{
  "Message": "Chat finished successfully"
}
```

#### Failed response body
##### 404 Not found
```json
{
  "ErrorDescription": "could not finish chat in salesforce : This contact does not have an interconnection"
}
```

### Get context

This resource gets the context according to the userID sent.

It can be used to validate that we are saving the chat with the bot in Redis and the webhook is receiving the messages.

`GET /v1/context/{{user_id}}`

#### Required role 

***YALO_ROLE***

#### Path params

| Name | Value | Required |
| :--- | :--- | :--- |
| user_id | `521125486585` or `facebookId` | Y |

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Response body 

##### 200 STATUS OK

```text
"Cliente [05-10-2021 16:41:39]:\n\nCliente [05-10-2021 18:43:59]:Restart\n\nBot [05-10-2021 18:44:00]:Puedo ayudarte a: \n\n*1.* Ubicar tu *tienda* más cercana\n*2.* Solicitar un *préstamo*\n*3.* Realizar un *abono*\n*4.* Resolver *dudas*\n*5.* Dar seguimiento a un *pedido*\n*6*. Recibir asistencia *sobre un producto*.\n\nEscribe el número de la opción o la palabra resaltada.\n\nCliente [05-10-2021 18:44:04]:1\n\nBot [05-10-2021 18:44:04]:Sigo aprendiendo como responder a tu solicitud. 😬\n\nBot [05-10-2021 18:44:06]:\nEscribe *Inicio* para volver al Inicio\nEscribe *Ayuda* para comunicarte con un asesor. \n\nCliente [05-10-2021 18:44:10]:Inicio\n\nBot [05-10-2021 18:44:10]:¡Hola, Fernando! Soy Coppelbot, 🤖 tu asistente virtual de *Coppel*, ¡es un gusto poder ayudarte y responder tus consultas por WhatsApp!\n\nBot [05-10-2021 18:44:11]:\nAntes de continuar, conoce nuestro *Aviso de privacidad*: http://bit.ly/coppelprivacidad\n\nBot [05-10-2021 18:44:12]:Puedo ayudarte a: \n\n*1.* Ubicar tu *tienda* más cercana\n*2.* Solicitar un *préstamo*\n*3.* Realizar un *abono*\n*4.* Resolver *dudas*\n*5.* Dar seguimiento a un *pedido*\n*6*. Recibir asistencia *sobre un producto*.\n\nEscribe el número de la opción o la palabra resaltada.\n\nCliente [05-10-2021 18:44:22]:1\n\nBot [05-10-2021 18:44:22]:Sigo aprendiendo como responder a tu solicitud. 😬\n\nBot [05-10-2021 18:44:25]:\nEscribe *Inicio* para volver al Inicio\nEscribe *Ayuda* para comunicarte con un asesor. \n\nCliente [05-10-2021 18:44:31]:Inicio\n\n"
```

#### Failed response body
##### 400 BAD REQUEST
```json
{
  "ErrorDescription": "Missing param : user_id"
}
```

### Register Webhook 

This resource allows us to register the webhooks mentioned above in ***integrations-api*** so that ***integrations-channels*** send requests to these endpoints and we receive the messages between the users and the bots.

***Note***: It must be taken into account that the registration through this endpoint is currently with version 1, so when registering a webhook for Facebook in prod that is through the ***passthrought*** channel it is necessary to register it in the DB as version 3, for what we need to request the support of the Platform team to set the update to version 3 directly in the database after using the endpoint.

`POST /v1/integrations/webhook/register/{{provider}}`

#### Required role

***YALO ROLE***

#### Path params

| Name | Value | Required |
| :--- | :--- | :--- |
| provider | `whatsapp` or `facebook` | Y |

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Response body

##### 200 Status

```json
{
  "Message": "Register webhook success with provider : whatsapp"
}
```

#### Failed response body
##### 500 Internal Server
```json
{
  "ErrorDescription": "error register webhook"
}
```

### Remove Webhook

This resource allows us to remove the webhooks mentioned above in integrations-api so that integrations-channels no longer send requests to these endpoints.

`DELETE /v1/integrations/webhook/remove/{{provider}}`

#### Required role

***YALO ROLE***

#### Path params

| Name | Value | Required |
| :--- | :--- | :--- |
| provider | `whatsapp` or `facebook` | Y |

#### Request header

| Name | Value | Required |
| :--- | :--- | :--- |
| Authorization | `Bearer ${token}` | Y only if token is not sent as queryParam |

#### Query params

| Name | Value | Required |
| :--- | :--- | :--- |
| token | `${token}` | Y only if the token is not sent in the Authorization header  |

#### Response body

##### 200 Status

```json
{
  "Message": "Remove webhook success with provider : whatsapp"
}
```

#### Failed response body
##### 500 Internal Server
```json
{
  "ErrorDescription": "error remove webhook"
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

  