#!/bin/sh
FILE="configs/${ENVIRONMENT}.env.yaml"

IFS=$'\n' read -d '' -r -a ARRAY < ${FILE} unset IFS

for i in ${!ARRAY[@]}; do
    ITEM=${ARRAY[$i]}

    KEY="${ITEM%%: *}"
    VALUE="${ITEM#*: }"
    ESCAPED_VALUE="${VALUE//'"'}"

    ENV_VARS+=${KEY}=${ESCAPED_VALUE}${GOOGLE_DELIMITER:-;}
done

echo ${ENV_VARS}
