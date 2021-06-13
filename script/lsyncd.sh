cat > config_$1 << EOF
    settings {
        inotifyMode = "CloseWrite",
        maxProcesses = 20,
        insist = true,
        nodaemon = true
    }

    sync {
        default.rsync,
        source    = "$2/",
        target    = "$1:$2/",
        delay     = 2,
        rsync     = {
            temp_dir = "/tmp/",
            rsh      = "sshpass -p root ssh  -o 'StrictHostKeyChecking no'"
        }
    }
EOF

# Updated note: sync only with primary replica

# Note: Current workaround to set the delay to 3 second, this to prevent
# weird behavior for bidirectional sync. Make sure your file can be fully written
# in 3 second, else increase the delay (slower replication)

/usr/bin/lsyncd -nodaemon -log all config_$1