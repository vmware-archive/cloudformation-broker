#!/bin/bash

./cloudformation-broker --config=config-sample.json

####################################################################################################################################

# Catalog
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/catalog"

####################################################################################################################################

# Provision CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}'

# Last Operation CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/service_instances/testcf/last_operation"

# Bind CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf/service_bindings/cf-binding" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722"}'

# Unbind CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testcf/service_bindings/cf-binding?service_id=91a53118-16fd-4865-b273-0257dbea0fe8&plan_id=fa5b59d8-e137-48b8-b696-6100d9cea722"

# Update CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"AccessControl":"PublicRead"}}'

# Last Operation CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/service_instances/testcf/last_operation"

# Deprovision CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true&service_id=91a53118-16fd-4865-b273-0257dbea0fe8&plan_id=fa5b59d8-e137-48b8-b696-6100d9cea722"

# Last Operation CloudFormation
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/service_instances/testcf/last_operation"

####################################################################################################################################

# Provision Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id","parameters":{"((("}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"unknown","organization_guid":"organization_id","space_guid":"space_id"}'

# Update Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testcf" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"},"parameters":{"((("}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"unknown","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/testcf?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"unknown","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PATCH "http://username:password@localhost:3000/v2/service_instances/unknown?accepts_incomplete=true" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","previous_values":{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722","organization_guid":"organization_id","space_guid":"space_id"}}'

# Deprovision Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/testcf?service_id=91a53118-16fd-4865-b273-0257dbea0fe8&plan_id=fa5b59d8-e137-48b8-b696-6100d9cea722"
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X DELETE "http://username:password@localhost:3000/v2/service_instances/unknown?accepts_incomplete=true&service_id=91a53118-16fd-4865-b273-0257dbea0fe8&plan_id=fa5b59d8-e137-48b8-b696-6100d9cea722"

# Bind Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/testcf/service_bindings/cf-binding" -d '{"service_id":"unknown","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722"}'
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X PUT "http://username:password@localhost:3000/v2/service_instances/unknown/service_bindings/cf-binding" -d '{"service_id":"91a53118-16fd-4865-b273-0257dbea0fe8","plan_id":"fa5b59d8-e137-48b8-b696-6100d9cea722"}'

# Last Operation Errors
curl -H 'Accept: application/json' -H 'Content-Type: application/json' -X GET "http://username:password@localhost:3000/v2/service_instances/unknown/last_operation"

####################################################################################################################################
