# Price Sync Calculation
Background
----------------
This service is used to calculate price sync related price.

You can find IDL definition and SPEX details from [Space API Namespace](https://space.shopee.io/spex/api_namespaces/api_namespace_categories/422010/api_namespaces/422089).

# Code Framework
Due to legacy issue, now price-sync-calculation has two code framework.
1. For global discount / oversea, processor -> dm (calc & data) -> service -> db/cache.
2. [Recommend] For other, processor -> logic -> service / repository -> cache.

We will try to migrate to use 2nd one when free.
So for new changes, pls use 2nd one also.
https://confluence.shopee.io/display/SPPT/%5BTRD%5D+Switch+to+use+jingyu%27s+framework+in+price+sync+calculation

# How to run/test/debug the service in local env?
Basic guide can refer [here](https://spkit.shopee.io/guide/tutorials/1.setup.html#run).
1. Make sure to [map remote Spex sockets for local testing](https://confluence.shopee.io/display/SPDC/Map+remote+Spex+sockets+for+local+testing), otherwise would get error like `error="dial unix .../spex.sock: connect: connection refused"`.
2. Build server. The easy cmd is `make/all`, and if u alr have downloaded deps and generated auto-gen codes, then can run `make/build` directly.
3. Run service. Just use `./bin/server` to run it.
4. Call your service locally via curl or postman. Use actual values of the following variables for your need.
- INSTANCE_ID = 612161b4e8754c859d578c850e942fd1 // For generated instance id. Would be different EVERY TIME after deployed.
- Api Param: need check log if register command with param. If yes, remember to add `?param=XXXX` after the command. Can check log via `routing_rules:\"price.sync_price.calculation.calc_global_discount_info_by_item_ids?=ebff6fdc7f61c2d0f5e6f114385963af161ceb23f394dbe5fc78e96c328de838\"`
- SDU = syncprice.calculation.global.test.master.default // For test env and default sdu, and can update in `etc/service.yml`. For testing and debugging, pls don't use default sdu, then requests of default envs won't be indirect into local env.
- SERVICE_KEY = a67df1b5099df4016bb05a675eee86e5 // For non live service key.
- CID: since this service is global service, so just use `global` in cid field.
```
curl --location --request POST 'https://http-gateway.spex.test.shopee.sg/sprpc/price.sync_price.calculation.calc_global_discount_info_by_item_ids?param=ebff6fdc7f61c2d0f5e6f114385963af161ceb23f394dbe5fc78e96c328de838' \
--header 'Content-Type: application/json' \
--header 'x-sp-servicekey: a67df1b5099df4016bb05a675eee86e5' \
--header 'x-sp-sdu: syncprice.calculation.global.test.master.default' \
--header 'x-sp-destination: syncprice.calculation.global.test.master.default.612161b4e8754c859d578c850e942fd1' \
--header 'shopee-baggage: CID=global' \
--header 'Cookie: SPC_EC=-; SPC_F=lEDRIViZ8rDJbYpjnTMAmbBaueDI4zdF; SPC_R_T_ID="aVBinT5PRGwS78iObQsD0G/gh//zuA82WZSsEh8PAHl3SFbqdGIzCZDta/WiO4U1kvWyv1eGe4RWbFKgWsUkT2SpRfNQhj0/Id+SnHyeswo="; SPC_R_T_IV="Zy9n2g3g49MD9HyShAbxKA=="; SPC_U=-' \
--data-raw '{
}'
```
**Notes**
- For the old version of spkit, we have `./bin/testserver` to do local test and debug,
  and due to not register apis, then requests of actual envs won't be indirect into local env.

  For the new version, you can update sdu_id in `etc/service.yml`, then it won't affect actual envs also.
