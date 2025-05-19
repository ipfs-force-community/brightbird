/app/droplet net listen 
if [ $? -eq 0 ]; then
    echo "droplet net listen success"
else
    echo "droplet net listen failed"
    exit 1
fi