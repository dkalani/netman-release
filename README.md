# netman-release

A [garden-runc](https://github.com/cloudfoundry-incubator/garden-runc-release) add-on
that provides container networking.

## kicking the tires on the policy server
0. push a couple test apps if you don't already have them:

  ```
  (cd src/netman-cf-acceptance/example-apps/proxy && cf push test1 && cf push test2)

  go install cf-cli-plugin && CF_TRACE=true cf uninstall-plugin connet; cf install-plugin -f bin/cf-cli-plugin && cf plugins

  cf net-allow test1 test2
  cf net-list
  cf net-disallow test1 test2
  cf net-list
  ```


## Deploy and test with Diego

Clone the necessary repositories

```bash
pushd ~/workspace
  git clone https://github.com/cloudfoundry-incubator/diego-release
  git clone https://github.com/cloudfoundry/cf-release
  git clone https://github.com/cloudfoundry-incubator/netman-release
popd
```

Run the deploy script

```bash
pushd ~/workspace/netman-release
  ./scripts/deploy-to-bosh-lite
popd
```

Finally, run the acceptance errand:

```bash
bosh run errand netman-cf-acceptance
```

## Deploy and test in isolation

```bash
bosh target lite

cd ~/workspace/netman-release

./scripts/update
bosh -n create release --force && bosh -n upload release --rebase
bosh deployment bosh-lite/deployments/netman-bare.yml

bosh -n deploy
bosh run errand acceptance-tests
```

## Running unit tests

```bash
docker-machine create --driver virtualbox --virtualbox-cpu-count 4 --virtualbox-memory 2048 dev-box
eval $(docker-machine env dev-box)
~/workspace/netman-release/scripts/docker-test
```

