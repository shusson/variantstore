# variantstore
Minimal variant store API layer (no samples or genotypes).

Connects to a mysql server loaded with variant information.

### Deployment

Containerized App available at:
https://hub.docker.com/r/shusson/variantstore

#### Example deployment using docker-compose

See [docker/.env](docker/.env) for deployment environment configuration

Start mysql server
```bash
docker-compose up -d sql
```

Optional - Load data
```bash
docker-compose up load
```

Start the api server
```bash
docker-compose -d up api
```

#### Backing up sql data

```bash
docker run --rm -v dockervs_data-volume:/tmp -v $(pwd):/backup ubuntu tar cvf /backup/backup.tar /tmp
```
