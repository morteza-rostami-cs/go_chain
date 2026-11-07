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

```
