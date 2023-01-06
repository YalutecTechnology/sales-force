# Interconnection #

Interconnection is an struct that is stored in Redis used to mantain an active session that represents the chat between a user (through a Yalo bot) and a live agent (through the Salesforce Console). This struct contains all the required information to be able to call the [Long Polling in Salesforce.](https://developer.salesforce.com/docs/atlas.en-us.live_agent_rest.meta/live_agent_rest/live_agent_rest_http_long_polling_loop.htm)

In the following table, the needed information that also is stored in Redis is shown:

### Interconnection Redis
| Field         | Type     | Description                                                                                           |
|---------------|----------|-------------------------------------------------------------------------------------------------------|
| userId        | string   | User identifier, on WA is the number with +5255... on Facebook is the facebookId                      |
| sessionId     | string   | Salesforce session identifier                                                                         |
| sessionKey    | string   | Salesforce session key                                                                                |
| affinityToken | string   | Salesforce token chat                                                                                 |
| status        | string   | Chat status: Failed, OnHold, Active, Closed                                                           |
| timestamp     | datetime | timestamp                                                                                             |
| provider      | string   | Message origin : Whatsapp or Messeger                                                                 |
| botSlug       | string   | Bot name : coppel-wa-staging                                                                          |
| botId         | string   | Bot's phone number                                                                                    |
| name          | string   | User´s Nickname                                                                                       |
| email         | string   | User's Email used to find or create a contact on Salesforce                                           |
| phoneNumber   | string   | User's phone number used to find or create a contact on Salesforce                                    |
| caseId        | string   | Optional: salesforce case id, could be used in the future to get and send the case number to the user |
| extraData     | hash     | To store additional information                                                                       |

### Interconnection

This struct represents the connection between a user and an agent into the Salesforce integration, is also an extension of the `Interconnection Redis`, but it stores additional information that allows to manage operations between both libraries, it will stores in a manager's map which is named as `InterconnectionMap[string]*Interconnection`.

| Field                     | Type        | Description                                                                                           |
|---------------------------|-------------|-------------------------------------------------------------------------------------------------------|
| id                        | string      | Redis key with the following suggested structure: {userid}-{sessionid}                                |
| userId                    | string      | User identifier, on WA is the number with +5255... on Facebook is the facebookId                      |
| sessionId                 | string      | Salesforce session identifier                                                                         |
| sessionKey                | string      | Salesforce session key                                                                                |
| affinityToken             | string      | Salesforce token chat                                                                                 |
| status                    | string      | Chat status: Failed, OnHold, Active, Closed                                                           |
| timestamp                 | datetime    | timestamp                                                                                             |
| provider                  | string      | Message origin : Whatsapp or Messeger                                                                 |
| botSlug                   | string      | Bot name : coppel-wa-staging                                                                          |
| botId                     | string      | Bot's phone number                                                                                    |
| name                      | string      | User´s Nickname                                                                                       |
| email                     | string      | User's Email used to find or create a contact on Salesforce                                           |
| phoneNumber               | string      | User's phone number used to find or create a contact on Salesforce                                    |
| caseId                    | string      | Optional: salesforce case id, could be used in the future to get and send the case number to the user |
| extraData                 | hash        | To store additional information                                                                       |
| salesforceMessageChannel  | channel     | Message channel to Salesforce                                                                         |
| integrationMessageChannel | channel     | Message channel to integrations                                                                       |
| agentID                   | string      | Current agent identifier                                                                              |
| numMessage                | int         | Message number, it would help to create an identifier from the message to be sent.                    |
| EventsChat                | listObjects | Long polling events from salesforce                                                                   |