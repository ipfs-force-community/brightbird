echo "set wallet password first ..."
password="123456"
printf "%s\n%s\n" "$password" "$password" | /app/venus wallet set-password
if [ $? -eq 0 ]; then
    echo "set wallet password success"
else
    echo "set wallet password failed"
    exit 1
fi

# check new wallet
DEFAULT_ADDR=$(/app/venus wallet new)
if [ $? -ne 0 ]; then
    echo "Failed to create wallet use default type" >&2
    exit 1
fi
echo "new default type address: $DEFAULT_ADDR"

BLS_ADDR=$(/app/venus wallet new --type bls)
if [ $? -ne 0 ]; then
    echo "Failed to create wallet use bls type" >&2
    exit 1
fi
echo "new BLS type address: $BLS_ADDR"

/app/venus wallet ls | grep -Fq "$DEFAULT_ADDR"
DEFAULT_ADDR_PASS=$?
/app/venus wallet ls | grep -Fq "$BLS_ADDR"
BLS_ADDR_PASS=$?

if [ $DEFAULT_ADDR_PASS -ne 0 ]; then
    echo "default type address [$DEFAULT_ADDR] not found"
    exit 1
elif [ $BLS_ADDR_PASS -ne 0 ]; then
    echo "BLS type address [$BLS_ADDR] not found"
    exit 1
else
    echo "ALL address found"
    exit 0
fi
