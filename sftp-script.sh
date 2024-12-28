
# Execute SFTP commands to get a list of dumps
sshpass -p "J7Beuv0YI9qQCMY" sftp -oPort=30222 -i "/root/.ssh/Olex.pem" jens@116.202.11.250 <<EOF
ls /data/
bye
EOF
# After getting the list, process it locally
latest_dump=$(ls /data/ | grep 'dump_' | sort -r | head -n 1)
echo "Latest dump: $latest_dump"
