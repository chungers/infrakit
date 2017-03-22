# set HTTP Proxy settings if set.
# make changes before docker is restarted.
if [ -n "$HTTP_PROXY" ]; then
    echo $HTTP_PROXY | mobyconfig set proxy/http
fi

if [ -n "$HTTPS_PROXY" ]; then
    echo $HTTPS_PROXY | mobyconfig set proxy/https
fi

if [ -n "$NO_PROXY" ]; then
    echo $NO_PROXY | mobyconfig set proxy/exclude
fi
