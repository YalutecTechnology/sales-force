test:
	./.yalo/utils/test-go.sh

# Mocks should be regenerated each time an interface changes
gen-mocks:
	# This is an example on how to generate mocks for several specific interfaces.
	# If you need to automatically regenerate mocks for other interfaces, use
	# the same command syntax and add it to this make directive
	# mocks for app/services
	mockery --name=SfcChatInterface \
		--dir base/clients/chat/ \
		--output app/services/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=SaleforceInterface \
		--dir base/clients/salesforce/ \
		--output app/services/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=SfcLoginInterface \
		--dir base/clients/login/ \
		--output app/services/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for app/manage
	mockery --name=SalesforceServiceInterface \
		--dir app/services/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=BotRunnerInterface \
		--dir base/clients/botrunner/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=StudioNGInterface \
		--dir base/clients/studiong/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=Producer \
		--dir base/subscribers/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=IInterconnectionCache \
		--dir base/cache/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=IContextCache \
		--dir base/cache/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=IMessageCache \
		--dir base/cache/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=IntegrationInterface \
		--dir base/clients/integrations/ \
		--output app/manage/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for app/cron
	mockery --name=SalesforceServiceInterface \
		--dir app/services/ \
		--output app/cron/mocks/ \
		--outpkg mocks \
		--case underscore
	mockery --name=IContextCache \
		--dir base/cache/ \
		--output app/cron/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/chat
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/chat/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/salesforce
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/salesforce/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/botrunner
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/botrunner/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/integrations
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/integrations/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/login
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/login/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for base/clients/studiong
	mockery --name=ProxyInterface \
		--dir base/clients/proxy/ \
		--output base/clients/studiong/mocks/ \
		--outpkg mocks \
		--case underscore
	# mocks for app/api/handlers
	mockery --name=ManagerI \
		--dir app/manage/ \
		--output app/api/handlers/mocks/ \
		--outpkg mocks \
		--case underscore
