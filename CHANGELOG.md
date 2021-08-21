
<a name="v0.0.1"></a>
## v0.0.1 (2021-08-21)

### Feat

* **Integration Enviroment:** Create skaffold configuration for stagging. Closes #CSF-23
* **Integration-Backbone:** Create main structure. Add project  base files : in-memory store library, proxy library, add endpoint healtcheck. Closes #CSF-43
* **Integration-Backbone:** Service to send message to Salesforce. Add endpoint to send a text message to Salesforce API. Closes #CSF-59
* **Integration-Backbone:** Service to send images to Salesforce. Add endpoints to upload and associate an image to a case. Closes #CSF-60
* **Integration-Backbone:** requester  to reconnect session for chat in  Salesforce. Add service to reconect session. Closes #CSF-56
* **Integration-Library Salesforce:** Set request to create chat sessions from Salesforce. Add request for create chat and tests. Closes #CSF-55
* **Integration-Library Salesforce:** Service to long polling messages from Salesforce. Add the request to obtain the agents' messages  of a chat, the security token is added to the request to obtain token. Closes #CSF-58
* **Integration-Library Salesforce:** Service to retrieve contact from Salesforce. Add request for retrieve contactID from Salesforce and tests. Closes #CSF-151

