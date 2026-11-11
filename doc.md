```bash

# clone only one branch
git clone --branch c1_genesis_json https://github.com/web3coach/the-blockchain-bar.git

# list available branches
git ls-remote --heads https://github.com/web3coach/the-blockchain-bar.git

# local branches
git branch

# bring down more branches
git fetch origin lesson-01
git switch -c lesson-01 origin/lesson-01

# switch between local branches
git switch main
git switch lesson-01

#============

# make a go module
go mod init github.com/<your-github-username>/the-blockchain-bar

mkdir -p cmd/tbb

# install cobra
go get github.com/spf13/cobra@latest
go mod tidy # remove unused packages

# build cobra commands, into:: $GOPATH/bin
go install ./cmd/tmc/...

## if tmc , did not work, do these::
go env GOPATH
/home/apax/go
> go env GOBIN
/home/apax/go/bin
>

ls /home/apax/go/bin/
dlv  gopls  helloworld  tmc

ls "$(go env GOPATH)/bin"
dlv  gopls  helloworld  tmc


# add Go bin dir to PATH
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
#================= end:: fixing tmc

# rebuild cmd commands each time you change something in ./cmd/tmc
go install ./cmd/tmc

tmc tx add --from=babayaga --to=caesar --value=1000
tmc tx add --from=babayaga --to=morteza --value=50
tmc tx add --from=morteza --to=morteza --value=100 --data=reward

# =============
# seed block.db
go run ./cmd/tmc/tmcmigrate/main.go

# define any pass and it creates , genesis.json and block.db
tmc balances list --datadir=./data

# if needed later, move the /data folder outside of project folder
tbb balances list --datadir=$HOME/data

#============
# http routes
curl -s -X GET http://localhost:8081/balances/list

curl -s -X POST http://localhost:8081/tx/add \
  -H "Content-Type: application/json" \
  -d '{"from":"morteza","to":"babayaga","value":100}'

# ===
# we can run command like this: and create our data in HOME
# use single quote, around path
tmc balances list --datadir='~/tmc-test'
tmc balances list --datadir='$HOME/tmc-test'
# path cleaning also works, /home/user/tbb-test
tmc balances list --datadir='$HOME/tmc-test/../tmc-test'

# inside project/data
tmc balances list --datadir=./data

```
