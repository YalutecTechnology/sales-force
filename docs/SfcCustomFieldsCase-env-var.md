
# Envar SfcCustomFieldsCase

This variable represents a string map that will contain all the custom fields of a customer in the Salesforce console to create their cases and that our service can send through the endpoint to create chats. An example of this package is the following:

```yaml
- name: SALESFORCE-INTEGRATION_SFC_CUSTOM_FIELDS_CASE
  value: "address:CP_Address__c,order_traking_id:CP_OrderTrackingId__c,source_flow_bot:CP_Source_Flow_Bot__c"
```

The previous values indicate that the ***"address"*** parameter will be received in the request to create a chat and it will be sent in the case of Salesforce as ***"CP_Address__c".***

For example, in the case of Coppel, in their cases they asked us to send the **address**, a value called **order tracking** and **origin of the bot flow**.

To do this, in the endpoint of creating Salesforce case, the following parameters must be sent as **"CP_OrderTrackingId__c"**, **"CP_Address__c"** and **"CP_Source_Flow_Bot__c"**

So in the **/v1/chats/connect** endpoint we must send that information in the **extra data** map as in the following example:

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

As seen in the previous example in the extra data we receive **"source_flow_bot"**, **"address"** and **"order_traking_id"**.

Internally, this data is sent to Salesforce by the integration service with the values indicated in the send **"CP_OrderTrackingId__c"**, **"CP_Address__c"** and **"CP_Source_Flow_Bot__c"** and the Salesforce API appends them to the case as shown below:

![info-chat-yalo.png](/docs/images/info-chat-yalo.png)