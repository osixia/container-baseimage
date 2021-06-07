# Environment Files

`.env` files in the `environment` directory and any sub-directories are loaded before executing services lifecycle scripts (`startup.sh`, `process.sh`, and `finish.sh`) or entrypoint lifecycle pre-commands.
The variables they contain are defined as environment variables in the container.

Files are loaded in alphabetical order, and variables are overwritten if they were already defined in a previous file.

**Container environment variables set at runtime override values defined in `.env` files.**
