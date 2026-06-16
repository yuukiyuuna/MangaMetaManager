#!/bin/sh

# Get environment variables, default to 1000 if not specified
USER_ID=${PUID:-1000}
GROUP_ID=${PGID:-1000}

# Function to fix permissions
fix_permissions() {
    echo "Adjusting permissions for UID=$USER_ID, GID=$GROUP_ID"

    # Create group if it doesn't exist
    if ! grep -q "^mmm:" /etc/group; then
        addgroup -g "$GROUP_ID" mmm
    fi

    # Create user if it doesn't exist
    if ! grep -q "^mmm:" /etc/passwd; then
        adduser -u "$USER_ID" -G mmm -h /app -D mmm
    fi

    # Ensure directories exist
    mkdir -p /app/data /app/logs

    # Fix ownership for persistent directories
    # We want the user to be able to write to these
    chown -R mmm:mmm /app/data /app/logs
    
    # Also ensure the web dist is readable (it should be, but just in case)
    chown -R mmm:mmm /app/web/dist

    # Ensure the app directory itself is searchable/writable for mmm
    chown mmm:mmm /app
}

# Check if running as root
if [ "$(id -u)" = '0' ]; then
    fix_permissions
    # Use su-exec to drop privileges and run the command
    exec su-exec mmm:mmm "$@"
else
    # If already running as non-root (e.g., docker run --user), just execute
    # but try to create directories anyway
    mkdir -p /app/data /app/logs 2>/dev/null
    exec "$@"
fi
