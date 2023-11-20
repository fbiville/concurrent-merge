# Concurrent `MERGE`

First, spin up a Neo4j server:

```shell
docker run --rm \
    --env NEO4J_AUTH='neo4j/letmein!' \
    --env NEO4J_ACCEPT_LICENSE_AGREEMENT=yes \
    --publish=7687:7687 --publish=7474:7474 \
    --health-cmd "cypher-shell -u neo4j -p 'letmein!' 'RETURN 1'" \
    --health-interval 5s \
    --health-timeout 5s \
    --health-retries 5 \
    neo4j:5-enterprise
```

Then run at least two processes in "parallel":

```shell
go run main.go -uri=neo4j://localhost -password='letmein!' -goroutine-count=1000 &
go run main.go -uri=neo4j://localhost -password='letmein!' -goroutine-count=1000 &
```