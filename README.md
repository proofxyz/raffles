# Proof raffles

This public repository contains all the necessary data to reproduce the raffles conducted by PROOF.
All raffles are hence verifiable by anyone.

## Structure

The subdirectories contain the lists that we drew from, together with an `entropy` file that stores the random seed driving the raffles.
The entropy is derived from an upcoming block that was announced in advance, rendering us unable to manipulate the drawing.

## Performing raffles

### Install dependencies

```bash
go install github.com/divergencetech/ethier/ethier@latest
```

### Run the script

```bash
./raffle.sh
```
