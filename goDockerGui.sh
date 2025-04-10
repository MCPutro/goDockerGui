#!/bin/bash

# Cek status Docker
if systemctl is-active --quiet docker; then
    echo "Docker sedang berjalan."
    echo "Menjalankan docker-gui..."
    # Cek apakah proses docker-gui sudah berjalan
        if pgrep -f docker-ui > /dev/null; then
            echo "docker-ui sudah berjalan."
        else
            echo "Menjalankan docker-ui..."
            ./docker-ui & disown
        fi
else
    echo "Docker tidak berjalan. Silakan nyalakan Docker terlebih dahulu."
fi
