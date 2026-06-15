#!/bin/sh

# Get environment variables, default to 1000 if not specified
USER_ID=${PUID:-1000}
GROUP_ID=${PGID:-1000}

# Check if running as root (if started with --user, skip these steps)
if [ "$(id -u)" = '0' ]; then
    echo "Adjusting permissions: UID=$USER_ID, GID=$GROUP_ID"

    # Create group if it doesn't exist
    if ! getent group mmm >/dev/null; then
        addgroup -g "$GROUP_ID" mmm
    fi

    # Create user if it doesn't exist
    if ! getent passwd mmm >/dev/null; then
        adduser -u "$USER_ID" -G mmm -h /app -D mmm
    fi

    # Ensure directories exist and fix ownership
    mkdir -p /app/data /app/logs
    chown -R mmm:mmm /app/data /app/logs /app/web

    # Use su-exec to drop privileges and run the command
    exec su-exec mmm:mmm "$@"
else
    # If already running as non-root (e.g., docker run --user), just execute
    exec "$@"
fi
