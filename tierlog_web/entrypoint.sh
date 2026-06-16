#!/bin/sh
set -e

# If .env does not exist, copy from .env.example
if [ ! -f /var/www/html/.env ]; then
    cp /var/www/html/.env.example /var/www/html/.env
fi

# Ensure SQLite file is present and writable
touch /var/www/html/database/database.sqlite
chown www-data:www-data /var/www/html/database/database.sqlite

# Run migrations
php artisan migrate --force

# Generate APP_KEY if empty in .env and not provided via env var
if ! grep -q "APP_KEY=base64:" /var/www/html/.env && [ -z "$APP_KEY" ]; then
    php artisan key:generate --force
fi

# Cache configurations
php artisan config:cache
php artisan route:cache
php artisan view:cache

# Set permissions again to be sure
chown -R www-data:www-data /var/www/html/storage /var/www/html/bootstrap/cache /var/www/html/database
chmod -R 775 /var/www/html/storage /var/www/html/bootstrap/cache

# Execute the default container command
exec apache2-foreground
