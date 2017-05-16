# variantstore
Minimal variant store API layer (no samples or genotypes).

Connects to a mysql server loaded with variant information.

### Deployment

Containerized App available at:
https://hub.docker.com/r/shusson/variantstore/

#### Example setup
Start mysql server
```bash
docker run -v `pwd`:/data --name sql -e MYSQL_ROOT_PASSWORD=a -d mysql:5.7.18 --secure-file-priv=/data
```

Init and load data into mysql server... (currently a manual process)

Start variant store api and link it to the sql server
```bash
docker run -it -d --link sql:sql -p 8080:8080 --name vs shusson/variantstore -d 'root:password@tcp(sql:port)/db'
```

