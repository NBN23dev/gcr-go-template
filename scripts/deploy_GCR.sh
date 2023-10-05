#!/bin/sh

SERVICE_NAME=`bash scripts/get_service_name.sh`
SERVICE_ACCOUNT=`bash scripts/get_service_account.sh ${SERVICE_NAME}`
APP_VERSION=`bash scripts/get_last_version.sh`
ENV_VARS=`bash scripts/get_env_vars.sh`
DELIMITER="^${GOOGLE_DELIMITER:-;}^"

# Copy Dockerfile to root
cp build/package/Dockerfile .

# Build and push image
gcloud builds submit --tag gcr.io/${GOOGLE_PROJECT_ID}/${SERVICE_NAME}:${APP_VERSION} ./build/package

# Deploy image
gcloud run deploy ${SERVICE_NAME} --image gcr.io/${GOOGLE_PROJECT_ID}/${SERVICE_NAME}:${APP_VERSION} --set-env-vars ${DELIMITER}${ENV_VARS} --platform managed --region ${GOOGLE_REGION} --service-account ${SERVICE_ACCOUNT} --quiet

# Update traffic
gcloud run services update-traffic ${SERVICE_NAME} --to-latest --platform managed --region ${GOOGLE_REGION} --quiet