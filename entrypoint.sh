#!/bin/sh

# Get environment variables, default to 1000 if not specified
USER_ID=${PUID:-1000}
GROUP_ID=${PGID:-1000}

# Function to fix permissions (only possible if running as root)
fix_permissions() {
    echo "Adjusting permissions for UID=$USER_ID, GID=$GROUP_ID"

    # Create group if it doesn't exist
    if ! grep -q "^mmm:" /etc/group; then
        # If the GID is already taken, we might have issues, but let's try
        addgroup -g "$GROUP_ID" mmm 2>/dev/null || addgroup mmm
    fi

    # Create user if it doesn't exist
    if ! grep -q "^mmm:" /etc/passwd; then
        adduser -u "$USER_ID" -G mmm -h /app -D mmm 2>/dev/null || adduser -G mmm -h /app -D mmm
    fi

    # Ensure directories exist
    mkdir -p /app/data /app/logs /app/data/logs

    # Fix ownership for persistent directories
    # This is crucial when volumes are mounted
    chown -R "$USER_ID":"$GROUP_ID" /app/data /app/logs
    
    # Also ensure the web dist is readable
    chown -R "$USER_ID":"$GROUP_ID" /app/web/dist

    # Ensure the app directory itself is searchable/writable
    chown "$USER_ID":"$GROUP_ID" /app
}

# Check if running as root
if [ "$(id -u)" = '0' ]; then
    fix_permissions
    # Use su-exec to drop privileges and run the command
    # We use the numeric IDs to be safe
    exec su-exec "$USER_ID":"$GROUP_ID" "$@"
else
    # If already running as non-root (e.g., docker run --user 1000:1000), 
    # we can't chown, but we can try to create subdirectories
    mkdir -p /app/data/logs 2>/dev/null
    exec "$@"
fi
