#! /bin/sh

case "$1" in
    -m)
      swag init --instanceName=marketplace -g api/marketplace/router.go -o docs/marketplace
      ;;
    *)  
      echo "$1 is not an option"
      ;;
esac
