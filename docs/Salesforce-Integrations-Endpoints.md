# Salesforce Integrations Endpoints #

## CREATE A CHAT BETWEEN YALO AND SALESFORCE

This endpoint will be invoked by a Yalo Studio lambda or Webhook when the Yalo bot user (WhatsApp or Facebook) enters the option that corresponds to **Receive Human Assistance** or **Consult an advisor** and will create the chat with a Salesforce agent.

**Method** : POST

**URL**: **/v1/chats/connect**

### Query Params ###

| Name  | Value                                                                                                                                                                                                                                  |
|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| token | ${token} |

**Headers**: 

- **Content-Type :**  application/json
- **Authorization:** Bearer ${token}

Note: If the JWT is sent as Header Authorization it does not need to be sent as query param, or if it is sent as query param it does not need to be sent as header.

**Request Example Whatsapp**: 

```json
{
    "userID" : "5217331175599",
    "botSlug" : "coppel-bot",
    "botId" : "521554578545",
    "name" : "Eduardo Ochoa",
    "provider" : "whatsapp",
    "email" : "ochoa@example.com",
    "phoneNumber" : "5217331175599",
    "extraData" : {
        "source_flow_bot" : "SFB001",
        "address": "Calle Andres Figueroa 10 B, col. 20 de Noviembre",
        "order_traking_id": "1225222565"
    }
}
```

**Response**: 

**Status**: 200

```json
{
    "Message": "Chat created succefully"
}
```

**Request Example Facebook**: 

```json
{
    "userID" : "6199190013487246",
    "botSlug" : "coppel-msn-staging",
    "botId" : "233726951928154",
    "name" : "Eduardo Ochoa",
	    "provider" : "facebook",
    "email" : "ochoa@example.com",
    "phoneNumber" : "5217331175599",
    "extraData" : {
        "source_flow_bot" : "SFB001",
        "address": "Calle Andres Figueroa 10 B, col. 20 de Noviembre, Iguala de la Independencia, Gro.",
        "order_traking_id": "1245414242"
    }
}
```

**Response**: 

**Status**: 200

```json
{
    "Message": "Chat created succefully"
}
```

Status 400 : Bad Request 

```json
{
    "ErrorDescription": "Error validating payload : Key: 'ChatPayload.Email' Error:Field validation for 'Email' failed on the 'required' tag"
}
```

Status 401: Not Authorization

```json
{
    "ErrorDescription": "An Authorization header or an Query Param with name token is required."
}
```

## Get JWT

This endpoint allows us to obtain a JWT and be able to use the resources of this integration service.

The JWT obtained must be sent as a query param or a header.
Note: The only resources that this token cannot access are webhooks registered in the integrations API.

**Method** : POST

**URL**: **/v1/authenticate**

**Request**: 

```json
{
  "username" : "${username}",
  "password" : "${password}"
}
```

**Response**: 

**200 HTTP Status:** 

```json
{
  "token": "${token}"  
}
```

**Status 401 Not Authorization:**

```json
{
  "ErrorDescription": "Invalid credentials."
}
```

## Validate JWT

This endpoint allows us to validate the JWT, returning the ROLE type of the JWT as a response.

**Method** : GET

**URL**: **/v1/tokens/check**

### Query Params ###

| Name  | Value                                                                                                                                                                                                                                  |
|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| token | ${token} |

**Headers**: 

- **Content-Type :**  application/json
- **Authorization:** Bearer ${token}

**Note**: If the JWT is sent as Header Authorization it does not need to be sent as query param, or if it is sent as query param it does not need to be sent as header.

**Response**: 

**200 HTTP Status:** 

```json
{
  "username": "${username}",
  "role" : "${role}"
}
```

**Status 403 FORBIDDEN:**

```json
{
  "ErrorDescription": "Invalid Authorization token."
}

```

## WhatsApp Webhook

This resource will be the webhook registered in API integrations through the following endpoint for the whatsapp bot:

```JSON
POST /api/{{channel}}/bots/{{botId}}/healthcheck
Headers
Authorization: Bearer ${token}
```
```JSON
{
	"phone": "+5215511223344",
  "webhook": "http://example.com/api",
	"media_token" : "12341234123"
}
```

Once registered in the previous endpoint, **integrations-channels** will send requests to this endpoint and we will receive the messages between the users and the WhatsApp bot.

We will do two things from the messages received.

1. We store the context of the user, with this we refer to the conversation between the bot and the user before requesting a chat. We only store this information for 24 hours.
2. If there is a chat between an agent and a user, the incoming messages are sent to the chat in Salesforce console.

**Method** : POST 
`POST /v1/integrations/whatsapp/webhook`

**URL**: **/v1/integrations/whatsapp/webhook**

**Headers**: 

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

### Example Document Payload

[JSON Editor Online - view, edit and format JSON online](https://jsoneditoronline.org/#right=local.fipape&left=cloud.62f3eb18ab1a41c38d9bdb6ca170f54a)

## Facebook Webhook ##
This resource will be the webhook registered in API integrations through the endpoint ***`/{{channel}}/bots/{{botName}}/healthcheck`*** for the Facebook bot.

Once registered in the previous endpoint, ***integrations-channels*** will send requests to this endpoint, and we will receive the messages between the users and the Facebook Bot.

We will do two things from the received messages.

1. We store the context of the user, by this we mean the conversation between the bot and the user before requesting a chat. We only store this information for 24 hours.
2. If there is a chat between an agent and a user, incoming messages are sent to the chat in the Salesforce agent console.

**Method** : POST

**URL**: **/v1/integrations/facebook/webhook**

**Headers**: 

**Request text message for passthrought channel**: 

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

**Request image message for passthrought channel**: 

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

For more details of the request sent by integrations see [here.](https://developers.facebook.com/docs/messenger-platform/webhooks)

### Example Document Payload

[JSON Editor Online - view, edit and format JSON online](https://jsoneditoronline.org/#right=local.fipape&left=cloud.f765e10892c84d0bac3e4ca8f0d80b3b)

**Response**: 

**200 HTTP Status:** 

```json
{
  "Message": "insert success"
}
```

**Status 403 FORBIDDEN:**

## Register Webhook

This resource allows us to register the webhooks mentioned above in ***integrations-api*** so that ***integrations-channels*** send requests to these endpoints and we receive the messages between the users and the bots.

***Note***: It must be taken into account that the registration through this endpoint is currently with version 1, so when registering a webhook for Facebook in prod that is through the ***passthrough*** channel it is necessary to register it in the DB as version 3, for what we need to request the support of the Platform team to set the update to version 3 directly in the database after using the endpoint.

**Method** : POST

**URL**: **/v1/integrations/webhook/register/{{provider}}**

**Path Params**:

- **provider**: The allowed values are ***facebook*** and ***whatsapp*** to register webhooks

**Response**: 

**200 HTTP Status:** 

```json
{
  "Message": "Register webhook success with provider : whatsapp"
}
```

**Status 500 Internal server error:**

```json
{
  "ErrorDescription": "error register webhook"
}

```

## Remove Webhook

This resource allows us to remove the webhooks mentioned above in ***integrations-api*** so that ***integrations-channels*** no longer send requests to these endpoints.

**Method** : DELETE

**URL**: **/v1/integrations/webhook/remove/{{provider}}**

**Path Params**:

- **provider**: The allowed values are ***facebook*** and ***whatsapp*** to remove webhooks.

**Response**: 

**200 HTTP Status:** 

```json
{
  "Message": "Remove webhook success with provider : whatsapp"
}
```

**Status 500 Internal server error:**

```json
{
  "ErrorDescription": "error remove webhook"
}

```

## Get Context

This resource gets the context according to the userID sent.

It can be used to validate that we are saving the chat with the bot in Redis and the webhook is receiving the messages.

**Method** : GET

**URL**: **/v1/context/{{userID}}**

**Path Params**:

- **userID**: prefix + phone number for whatsapp ex:`521125486585` or `facebookId`

**Response**: 

**200 HTTP Status:** 

```json
[
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646674958,
        "text": "¡Hola! Soy Coppelbot, 🤖 tu asistente virtual de *Coppel*, ¡es un gusto poder ayudarte y responder tus consultas por WhatsApp!\n",
        "from": "bot",
        "ttl": "2021-11-12T16:04:36.583809067Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646677100,
        "text": "Puedo ayudarte a: \n\n*1.* Solicitar ayuda para *comprar*\n*2.* Solicitar un *préstamo*\n*3.* Realizar un *abono*\n*4.* Resolver *dudas*\n*5.* Dar seguimiento a un *pedido*\n*6*. Ubicar tu *tienda* más cercana\n\n\n\nEscribe *solo el número de la opción* que prefieras.",
        "from": "bot",
        "ttl": "2021-11-12T16:04:38.27335764Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646680103,
        "text": "Muy bien, nos vamos a comunicar con un asesor.\n",
        "from": "bot",
        "ttl": "2021-11-12T16:04:42.973667136Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646697866,
        "text": "9",
        "from": "user",
        "ttl": "2021-11-12T16:04:59.383432191Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646700811,
        "text": "\nNo ingreses mayúsculas ni caracteres especiales.\n\nEjemplo\n✅ juanperez@coppel.com",
        "from": "bot",
        "ttl": "2021-11-12T16:05:02.364503418Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636646712031,
        "text": "¡Hola! En un momento un agente te atenderá. Te recuerdo nuestro horario de atención:\n\nLunes a viernes de 8:00 a.m. a 11:45 p.m.\nSábados y domingos, de 8:00 a.m. a 9:45 p.m. hora del Centro.",
        "from": "bot",
        "ttl": "2021-11-12T16:05:13.663290341Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636647463085,
        "text": "\nEscribe *Hola* para volver al menú principal.",
        "from": "bot",
        "ttl": "2021-11-12T16:17:44.773388045Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636671898780,
        "text": "Sigo aprendiendo como responder a tu solicitud. 😬\n",
        "from": "bot",
        "ttl": "2021-11-12T23:05:00.967862273Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636671948681,
        "text": "Ayuda",
        "from": "user",
        "ttl": "2021-11-12T23:05:50.973381509Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636671951355,
        "text": "\n¿Cuál es el motivo de tu consulta? Elige un número 👇🏼\n\n*1*. Quiero saber dónde está mi pedido\n*2*. Quiero encontrar o comprar un artículo \n*3*. Quiero mi Estado de Cuenta \n*4*. Quiero registrar o saber el estatus de mi queja \n*5*. Solicitar cancelación de mi pedido \n*6*. Solicitar servicio de garantía\n*7*. Solicitar información sobre mi abono \n*8*. Recibí una notificación y deseo más información \n*9*. Alguna otra duda ",
        "from": "bot",
        "ttl": "2021-11-12T23:05:53.487151062Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636671954569,
        "text": "Perfecto. 🙂\nAntes de contactar un asesor, por favor escribe tu correo electrónico.\n",
        "from": "bot",
        "ttl": "2021-11-12T23:05:56.176958099Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636671977828,
        "text": "eduardo.ochoam2593@gmail.com",
        "from": "user",
        "ttl": "2021-11-12T23:06:20.083482397Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636672994709,
        "text": "Gracias por tu tiempo. 🙂\n",
        "from": "bot",
        "ttl": "2021-11-12T23:23:17.073401818Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636673008941,
        "text": "Holaaaaa",
        "from": "user",
        "ttl": "2021-11-12T23:23:30.285571523Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636673009955,
        "text": "\nAntes de continuar, conoce nuestro *Aviso de privacidad*: http://bit.ly/coppelprivacidad\n\n",
        "from": "bot",
        "ttl": "2021-11-12T23:23:31.563353943Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636673122956,
        "text": "Hdhhdhdhdjdhdjdjej",
        "from": "user",
        "ttl": "2021-11-12T23:25:25.163241786Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636756917842,
        "text": "Hdhhdhdhdjdhdjdjej",
        "from": "user",
        "ttl": "2021-11-13T22:41:59.26337658Z"
    },
    {
        "userId": "5217331175599",
        "client": "coppel",
        "timestamp": 1636756918574,
        "text": "Por favor, intenta hacer tu *pregunta de otra forma*, o escribe *Asesor* para que un agente pueda asistirte. Para regresar al inicio escribe *Hola*.",
        "from": "bot",
        "ttl": "2021-11-13T22:42:00.165536271Z"
    }
]

```

**Status 500 Internal server error:**

## Reference:

[API README](/app/api/handlers/README.md)