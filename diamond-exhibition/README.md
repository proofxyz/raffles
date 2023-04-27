# Diamond Exhibition choice allocation

Although not strictly a raffle, we include this code here for verification of
fair allocations.

The binary accepts the ranked preferences for 21 artworks with limited supply,
and performs an optimisation to find the best possible stable allocation. The
loss function is the sum of the preference number allocated to each entrant.

There are many possible stable allocations—even many with the same loss
value—thus we rely on entropy out of our control to randomly initialise the
algorithm. We commit to the (future) Ethereum mainnet block,
[17137900](https://etherscan.io/block/17137900), and will use the block hash
as the aforementioned entropy.

Although a random shuffling without any optimisation would also produce a stable
allocation, our experiments show an improvement of ~2,750 points on the loss
function. This equates to each entrant, on average, being 0.81 positions better
off.
