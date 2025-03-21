#! /bin/sh

case "$1" in
    -m)
      swag init --instanceName=marketplace -g api/marketplace/router.go -o docs/marketplace --exclude=api/v1,api/admin
      ;;
    -v1)
      swag init --instanceName=api_v1 -g api/v1/router.go -o docs/v1 --exclude=api/marketplace,api/admin
      ;;
    -a)
      swag init --instanceName=admin -g api/admin/router.go -o docs/admin --exclude=api/marketplace,api/v1
      ;;
    *)  
      echo "$1 is not an option"
      ;;
esac
