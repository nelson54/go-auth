docker run -d \
  -p 5432:5432 \
	--name go_auth \
	-e POSTGRES_USER=user \
	-e POSTGRES_PASSWORD=pass \
	-e POSTGRES_DB=auth \
	-e PGDATA=/var/lib/postgresql/data/pgdata \
	-v ~/.docker_mounts/auth:/var/lib/postgresql/data \
	postgres:alpine