#! /bin/sh

case "$1" in
    -m)
      swag init --instanceName=marketplace -g api/marketplace/router.go -o docs/marketplace  --exclude api/crud,api/indexer,api/discordbot,api/admin,api/pumpfun
      ;;
    -a)
      swag init --instanceName=admin -g api/admin/router.go -o docs/admin  --exclude api/crud,api/indexer,api/discordbot,api/marketplace,api/pumpfun
      ;;
    *)  
      echo "$1 is not an option"
      ;;
esac
