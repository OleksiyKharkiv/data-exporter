LATEST_DUMP=$(sudo sshpass -p 'J7Beuv0YI9qQCMY' sftp -oPort=30222 jens@116.202.11.250 <<EOF | grep '^dump_' | sort -r | head -n 1
cd /data
ls -1
bye
EOF
)

sudo sshpass -p 'J7Beuv0YI9qQCMY' sftp -oPort=30222 jens@116.202.11.250 <<EOF
cd /data/$LATEST_DUMP
lcd /var/lib/mongodb/download/dump
get -r mats-payment
get -r mats-user
get -r mats-diagnostic
get -r mats-training-plan
bye
EOF
