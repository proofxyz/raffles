# Proof raffles

This public repository contains all the necessary data to reproduce the raffles conducted by PROOF.
All raffles are hence verifiable by anyone.

## Structure

The subdirectories contain the lists that we drew from, together with an `entropy` file that stores the random seed driving the raffles.
The entropy is derived from an upcoming block that was announced in advance, rendering us unable to manipulate the drawing.

## Reproducing raffles

You can either reproduce the raffles yourself on your local machine or have a look out the outputs of our [CI](https://github.com/proofxyz/raffles/actions/workflows/raffle.yml).

### Install dependencies

```bash
go install github.com/divergencetech/ethier/ethier@v0.35.3
```

### Run the script

```bash
./raffle.sh
```
