#!/bin/sh

if [ $# -eq 0 ]; then
    echo "Missing Argument: Service name" 1>&2
    exit 1
fi

PROJECT=nbn23-competition-manager

if [ "$ENVIRONMENT" = "develop" ]; then
    PROJECT=$PROJECT-dev
fi

echo "gcr-$1@$PROJECT.iam.gserviceaccount.com"