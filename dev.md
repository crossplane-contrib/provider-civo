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

## Testing the provider locally

You can take advantage of a guided script to set up a `crossplane-master` cluster in Civo, deploying the Crossplane civo provider and running the provider locally.
To init the `crossplane-master` in Civo you can:
```bash
./scripts/local.sh init
```
To update the provider CRs and deploy it locally:
```bash
./scripts/local.sh init
```
To show an example you can apply to `crossplane-master` and check if everything is working:
```bash
./scripts/local.sh example
```