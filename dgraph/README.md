# Dgraph

To update the schema using the schema.dql file:

```bash
curl -X POST https://<YOUR_DGRAPH_URI>/dgraph/schema \
    --header "Authorization: Bearer {{TOKEN}}" \
    --header "Content-Type: application/dql" \
    --data @schema.dql
```

To delete all data in your Dgraph instance (but keep the schema):

```bash
curl -X POST https://<YOUR_DGRAPH_URI>/dgraph/alter \
    --header "Authorization: Bearer {{TOKEN}}" \
    --header "Content-Type: application/json"  \
    --data '{"drop_op": "DATA"}'
```