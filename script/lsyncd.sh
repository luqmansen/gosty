cat > config$1 << EOF
    sync {
        default.rsync,
        source    = "$2/",
        target    = "$1:/$2/",
        delay     = 5,
        rsync     = {
            archive  = true,
            compress = true,
            update   = true,
            temp_dir = "/tmp",
            rsh      = "sshpass -p root ssh  -o 'StrictHostKeyChecking no'"
        }
    }
EOF

# Note: Current workaround to set the delay to 5 second, this to prevent
# weird behavior for bidirectional sync. Make sure your file can be fully written
# in 5 second, else increase the delay (slower replication)

/usr/bin/lsyncd -nodaemon -log scarce -insist config$1