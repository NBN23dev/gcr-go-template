import json, os, time

"""Reads the script configuration from a file
Returns:
    Script configuration
"""
def get_config():
    script_dir = os.path.dirname(__file__)

    script_cfg_file = open(script_dir + "/config.json")
    config = json.load(script_cfg_file)
    script_cfg_file.close()

    return config


"""Reads the env vars that must be linked to Google Cloud secrets
Returns:
    Array of strings
"""
def get_secrets() -> list:
    secrets = []
    secrets_filename = "configs/secrets.env.json"
    if os.path.isfile(secrets_filename):
        secrets_file = open(secrets_filename)
        secrets = json.load(secrets_file)
        secrets_file.close()

    return secrets


"""Gets a unique identifier to be used as part of the docker image name.
The function checks teh value of the env var ENVIRONMENT to determine if it
is creating an image for the production environment (ENVIRONMENT=production)
Returns:
    str: Unique identifier
"""
def get_service_version():
    if os.getenv("$ENVIRONMENT") == "production":
        return os.getenv("$GITHUB_REF").split("/")[-1]

    return int(time.time())


"""Gets the gcloud builds command to create the container
Returns:
    str: unique docker image name
"""
def get_docker_image_name():
    project_id = os.getenv("GOOGLE_PROJECT_ID")
    service_name = os.getenv("SERVICE_NAME")

    return f"""gcr.io/{project_id}/{service_name}:{get_service_version()}"""


"""Runs a system command.
If dryRun is set true in the script configuration, 
the function runs dry and just prints the command.
If debug is set to true in the script configuration,
the command is both printed and run.
Args:
    cmd (string): Command to be run
    config (any): Script configuration dictionary
Returns:
    int: exit code returned by the 'os.system(cmd)' command 
    or 0 if runnin dry 
"""
def run_command(cmd, config):
    if config["dryRun"] == True:
        print(cmd)
        return 0
    
    if config["debug"] == True:
        print(f"""Running command: {cmd}""")

    result = os.system(cmd)

    if config["debug"] == True:
        print(f"""Command execution completed with result value {result}""")

    return result


"""Requests gcloud to submit a docker image build
Args:
    docker_image_name (string): Name of the image being built
Returns:
    int: exit code returned by the 'gcloud builds' command
"""
def request_cloud_build(config, docker_image_name):
    git_username = os.getenv("GIT_USERNAME")
    git_token = os.getenv("GIT_TOKEN")

    cmd = f"""gcloud builds submit --config=build/package/cloudbuild.yaml --substitutions=_GIT_USERNAME={git_username},_GIT_TOKEN={git_token},_IMAGE={docker_image_name}"""
    
    result = run_command(cmd, config)

    return result


"""Requests gcloud to deploy a docker image and forwarding the traffic to it
Args:
    imageName (string): Image name
    config (any): Script configuration
    secrets (list): Array of env vars that must be linked to Google Cloud secrets
Returns:
    int: exit code returned by the 'gcloud deploy' command
"""
def request_cloud_deploy(imageName, config, secrets):
    service_name = os.getenv("SERVICE_NAME")
    app_env = os.getenv("ENVIRONMENT")
    region = os.getenv("GOOGLE_REGION")
    service_account = config["serviceAccounts"][app_env]

    env_vars_file = f"""{config["configsLocation"]}/{app_env}.env.yaml"""

    cmd = f"""gcloud run deploy {service_name} \\
    --image {imageName} \\
    --env-vars-file {env_vars_file} \\
    --platform managed \\
    --region {region} \\
    --service-account {service_account} \\
    --quiet \\
    """

    if len(secrets) > 0:
        set_secrets = ""
        for secret in secrets:
            set_secrets += f"{secret}={secret}:latest,"
        set_secrets = set_secrets[:-1]
        cmd += f"--set-secrets={set_secrets}"

    result = run_command(cmd, config)
    if result != 0:
        return result

    cmd = f"""gcloud run services update-traffic {service_name} \\
    --to-latest \\
    --platform managed \\
    --region {region} \\
    --quiet"""

    return run_command(cmd, config)


# Main

config = get_config()

if config["dryRun"]:
    print("Running in dry mode")
elif config["debug"]:
    print("Running in debug mode")

secrets = get_secrets()
if config["debug"]:
    print(f"""Secrets Obtained: {secrets}""")


image_name = get_docker_image_name()
if config["debug"]:
    print(f"""Docker image name: {image_name}""")

request_cloud_build(config, image_name)
request_cloud_deploy(image_name, config, secrets)

if config["debug"]:
    print("Script finished")