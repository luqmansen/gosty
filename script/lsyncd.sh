cat > config$1 << EOF
    sync {
        default.rsync,
        source    = "$2/",
        target    = "$1:/$2/",
        delay     = 0,
        rsync     = {
            archive  = true,
            compress = true,
            rsh      = "sshpass -p root ssh  -o 'StrictHostKeyChecking no'"
        }
    }
EOF

/usr/bin/lsyncd -nodaemon -log all -insist config$1