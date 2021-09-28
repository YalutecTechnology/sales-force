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
  "token": "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJwYXNzd29yZCI6ImEyZWY5N2RjNWQ3MWYzNjUwM2NjY2UzYTE3NjRmYmQyYzAyNjgxNDQiLCJyb2xlIjoiQURNSU5fUk9MRSIsInVzZXJuYW1lIjoiYWRtaW5Vc2VyIn0.2-CDkQwBwiAjsC5APT356Xn5TeZwSEG6LHL7Da8hBSK--qJVzMcKe4pUTbCNtlr8"  
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

### Webhook for whastapp Bot

This resource will be the webhook registered in API integrations through the endpoint [`https://api-staging2.yalochat.com/integrations/api/{{channel}}/bots/{{botName}}/healthcheck`](https://www.notion.so/yalo/Integrations-API-00643b2f689943bb9daa1c5d064f0b32#de3c9ae04829401e86f756092e0e58e2) for the whatsapp bot.

Once registered in the previous endpoint, [integrations-channels](https://www.notion.so/yalo/Integrations-Daemon-0e6812a890124f018c1f4123bec8f42f#653893087c7040b8ba096940f3f1433a) will send requests to this endpoint and we will receive the messages between the users and the WhatsApp bot.

We will do two things from the messages received.

1. We store the context of the user, with this we refer to the conversation between the bot and the user before requesting a chat. We only store this information for 24 hours.
2. If there is a chat between an agent and a user, the incoming messages are sent to the chat in Salesforce.

`POST /v1/integrations/whatsapp/webhook`

#### Required role 

***NOT REQUIRED***

#### Request header

**TODO: Add security documentation**

#### Request body
##### Example text message for outgoing channel
```json
{
  "from": "5215573735001",
  "id": "ABGHUhVXNzUALwIQZWnmjBi7IBmfp2EfNrYbRQ",
  "text": {
    "body": "Hola"
  },
  "timestamp": "1632367928574",
  "type": "text"
}     
```

##### Example image message for outgoing channel
```json
{
  "from": "5215514732842",
  "id": "ABGHUhVRRzKEPwIKOo4a1mKpR_2KoA",
  "image": {
    "caption": "",
    "mimeType": "image/jpeg",
    "url": "https://api-staging2.yalochat.com/hooks-proxy/whatsapp/coppel-wa-staging/media/a97f26b7-1ff4-4803-b3c0-3974b43f807d"
  },
  "timestamp": "1630107882351",
  "type": "image"
}     
```

More information about the request body [here](https://www.notion.so/yalo/Integrations-Daemon-0e6812a890124f018c1f4123bec8f42f#653893087c7040b8ba096940f3f1433a)

#### Response body 

##### 200 Status

```json
{
  "Message": "insert success"
}
```

#### Failed response body
##### Bad Request
```json
{
  "ErrorDescription": "invalid pauload received : error decode body request"
}
```

### Webhook for Facebook Bot

This resource will be the webhook registered in API integrations through the endpoint [`https://api-staging2.yalochat.com/integrations/api/{{channel}}/bots/{{botName}}/healthcheck`](https://www.notion.so/yalo/Integrations-API-00643b2f689943bb9daa1c5d064f0b32#de3c9ae04829401e86f756092e0e58e2) for the facebook bot.

Once registered in the previous endpoint, [integrations-channels](https://www.notion.so/yalo/Integrations-Daemon-0e6812a890124f018c1f4123bec8f42f#653893087c7040b8ba096940f3f1433a) will send requests to this endpoint and we will receive the messages between the users and the Facebook bot.

We will do two things from the messages received.

1. We store the context of the user, with this we refer to the conversation between the bot and the user before requesting a chat. We only store this information for 24 hours.
2. If there is a chat between an agent and a user, the incoming messages are sent to the chat in Salesforce.

`POST /v1/integrations/facebook/webhook`

#### Required role 

***NOT REQUIRED***

#### Request header

**TODO: Add security documentation**

#### Request body
##### Example text message for passthrought channel
```json
{
  "authorRole": "user",
  "botId": "messeger-staging-bot",
  "message": {
    "entry": [
      {
        "id": "233726951928054",
        "messaging": [
          {
            "message": {
              "mid": "m_rhq1CGI9iAqiwi8rcjFwG19GI3HsmSbOU_4gIL4157f_PzsdjOhs5dv7ak0bNJQn5Xeta4YupUzDgDTzcW03sg",
              "text": "Hola"
            },
            "recipient": {
              "id": "233726951928054"
            },
            "sender": {
              "id": "5018811078134037"
            },
            "timestamp": 1632348042261
          }
        ],
        "time": 1632348042540
      }
    ],
    "object": "page"
  },
  "msgTracking": {},
  "provider": "facebook",
  "timestamp": 1632348042540
}      
```

##### Example image message for passthrought channel
```json
{
  "authorRole": "user",
  "botId": "messeger-staging-bot",
  "message": {
    "entry": [
      {
        "id": "233726951928054",
        "messaging": [
          {
            "message": {
              "attachments": [
                {
                  "payload": {
                    "url": "https://scontent.xx.fbcdn.net/v/t1.15752-9/242696087_456158895579766_8081273236399441617_n.jpg?_nc_cat=105&ccb=1-5&_nc_sid=58c789&_nc_ohc=UZuikmiYkqkAX-S1dYK&_nc_ad=z-m&_nc_cid=0&_nc_ht=scontent.xx&oh=4b27eb86d58a352ed0efdcfc5979dcfb&oe=6171D4BF"
                  },
                  "type": "image"
                }
              ],
              "mid": "m_1Cc4RFq8tdQhvyOBwHdVrV9GI3HsmSbOU_4gIL4157c5PC2EYygD1-GWTMakPi5-3PD5kJdLqUERgXPWxE-DNQ"
            },
            "recipient": {
              "id": "233726951928054"
            },
            "sender": {
              "id": "5018811078130077"
            },
            "timestamp": 1632351515407
          }
        ],
        "time": 1632351515564
      }
    ],
    "object": "page"
  },
  "msgTracking": {},
  "provider": "facebook",
  "timestamp": 1632351515564
}
```

More information about the request body [Facebook Message - Integration](https://www.notion.so/yalo/Facebook-Message-Integration-877564eff1104ba2b3d5bd9f170eed69)

#### Response body 

##### 200 Status

```json
{
  "Message": "insert success"
}
```

#### Failed response body
##### Bad Request
```json
{
  "ErrorDescription": "invalid payload received : error decode body request"
}
```

### End Chat

Este recurso finaliza el chat de acuerdo al userID enviado. Solo si existe un chat con este userID será finalizado.

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

  