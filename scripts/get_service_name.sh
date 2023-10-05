#!/bin/sh
FILE="configs/${ENVIRONMENT}.env.yaml"

IFS=$'\n' read -d '' -r -a ARRAY < ${FILE} unset IFS

for i in ${!ARRAY[@]}; do
    ITEM=${ARRAY[$i]}

    KEY="${ITEM%%: *}"
    VALUE="${ITEM#*: }"

    if [ "$KEY" = "SERVICE_NAME" ]; then
        echo ${VALUE}
    fi
done
