# Salesforce-Integration - Buttons and Owners IDs #

The **buttonId** is an identifier required when requesting a chat with the Salesforce API, because this identifier is associated with a query queue, in which the agents assigned to that queue can receive the chats created with this **buttonId**.

The **ownerId** on the other hand is an identifier that can be a QueueId or a UserId, with this, this value can be associated with a case, this value is optional, we only send it if it is required by the client's requirements, in the case of Coppel if it was to associate a Department field with your cases.

The integration requires at least one buttonId to be configured, for this we need to use the **SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT** envar

**SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT** is a map that will help us choose the subject, the **butonId** and the **ownerId** that is required for the cases and direct to the corresponding queues according to the client's requirement.

You can configure a map with default values in case there is no more than one queue. For instance:

```yaml
- name: SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT
  value: 'default={"subject":" Caso creado por Yalo Bot","providers":{"whatsapp":{"button_id":"{{buttonId}}","owner_id":""},"facebook":{"button_id":"{{buttonId}}","owner_id":""}}}'
```

If there are two queues, but one for the chats created for the Facebook bot and another for the WhatsApp chats, but both have the same OwnerId, we can use the following example: 

```yaml
- name: SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT
  value: 'default={"subject":" Caso creado por Yalo Bot","providers":{"whatsapp":{"button_id":"{{buttonId_whatsAppQueue}}","owner_id":"{{ownerId}}"},"facebook":{"button_id":"{{buttonId_facebookQueue}}","owner_id":"{{ownerId}}"}}}'
```

If there are more queues depending on a set of options required by the client, we can also use to send **SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_FIELD**, this is a parameter that we will send in the extraData object, when making a request to the ***/v1/chats/*connect** endpoint, this parameter is by default **source_flow_bot**.

Let's imagine that the client requires these two options and each one requires that it be sent to different queues according to each bot.

## Copy of Ejemplo varias colas

| Option | Bot      | Subject           | ButtonId  | OwnerId  |
|--------|----------|-------------------|-----------|----------|
| YA001  | whatsApp | Caso con opción 1 | buttonId1 | ownerId1 |
| YA001  | facebook | Caso con opción 1 | buttonId2 | ownerId2 |
| YA002  | whatsApp | Caso con opción 2 | buttonId3 | ownerId3 |
| YA002  | facebook | Caso con opcion 2 | buttonId4 | ownerId4 |

So for this, in the request **/chats /connect**, one of the previous options must be sent in the extraData in the source_flow_bot parameter, example:So for this, in the request ** / chats / connect **, one of the previous options must be sent in the extraData in the source_flow_bot parameter, example:

```json
{ ...
	"extraData" : {
									"source_flow_bot" : "YA002"
								}
}
```

So for this, in the request **/chats/connect**, one of the previous options must be sent in the extraData in the ***source_flow_bot*** parameter, example:

```yaml
- name: SALESFORCE-INTEGRATION_SFC_SOURCE_FLOW_BOT
  value: 'YA001={"subject":"Caso con opción 1","providers":{"whatsapp":{"button_id":"buttonId1","owner_id":"ownerId1"},"facebook":{"button_id":"buttonId2","owner_id":"ownerId2"}}};YA002={"subject":"Caso con opción 2","providers":{"whatsapp":{"button_id":"buttonId3","owner_id":"ownerId3"},"facebook":{"button_id":"buttonId4","owner_id":"ownerId4"}}};default={"subject":"Caso creado por Yalo Bbot","providers":{"whatsapp":{"button_id":"buttonId1","owner_id":"ownerId1"},"facebook":{"button_id":"buttonId1","owner_id":"ownerId1"}}}'
```