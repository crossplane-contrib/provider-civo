## Development of provider

### Cluster setup
- Install/Update the CRDs in the cluster by:
`kubectl apply -f package/crds`
- Create the provider config(make sure to update your API key) by:
`kubectl apply -f examples/civo/provider`

### Generating CRDs
- Types are defined under `apis/civo`
- Make sure not to change Go data type of existing fields as that lead be a breaking change for existing users. 
- For introducing new fields, add the new field and run `make generate` to generate new CRDs

## Running the Provider
- Run `go run cmd/provider/main.go` to start running the local version of the provider