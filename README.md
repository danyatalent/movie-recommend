# repo for movie recommendation service

## Build and deploy

use bash script in deploy/


example:
```bash
bash deploy.sh -u danya -h 158.160.124.149 -r /home/danya/deploy/bin
```

running application on the server requires .env file and path to config

example:

```bash
bin/app -config-path configs/config.yaml
```
