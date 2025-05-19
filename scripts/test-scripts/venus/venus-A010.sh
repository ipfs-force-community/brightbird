# set password
password="123456"
printf "%s\n%s\n" "$password" "$password" | /app/venus wallet set-password
if [ $? -eq 0 ]; then
    echo "set wallet password success"
else
    echo "set wallet password failed"
    exit 1
fi
