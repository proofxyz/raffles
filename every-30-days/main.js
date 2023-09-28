const { Alchemy, Network } = require("alchemy-sdk");
const fs = require("fs");
require("dotenv").config({ path: "../../proof/projects/proof.xyz/.env" });

// E30D contract address
const address = "0x5ab44d97b0504ed90b8c5b8a325aa61376703c88";
const TOKEN_ID = 5;
const FILEPATH = "./participants";

const config = {
  apiKey: process.env.NEXT_PUBLIC_ALCHEMY_API_KEY,
  network: Network.ETH_MAINNET,
};
const alchemy = new Alchemy(config);

async function getEntriesByOwner() {
  // Get owners
  const response = await alchemy.nft.getOwnersForContract(address, {
    withTokenBalances: true,
  });

  // Map of `ownerAddress` to `how many of this token do they own`
  const entriesByOwner = {};
  response.owners.forEach((owner) => {
    const { ownerAddress, tokenBalances } = owner;
    // a tokenBalance is an object that maps { [tokenId]: numOwned }
    owner.tokenBalances.forEach((tokenBalance) => {
      const { tokenId } = tokenBalance;
      // tokenId is a 0x string
      if (parseInt(tokenId, 16) === TOKEN_ID) {
        const countOwned = tokenBalance.balance;
        entriesByOwner[ownerAddress] = countOwned;
      }
    });
  });

  return entriesByOwner;
}

const handleError = (err) => {
  if (err) console.error(err);
};

const main = async () => {
  const entriesByOwner = await getEntriesByOwner();

  // generate a gimongous array and turn it into a string connected by \n
  const fileContents = Object.entries(entriesByOwner)
    .reduce((prev, [owner, numEntries]) => {
      return [...prev, ...Array.from({ length: numEntries }, () => owner)];
    }, [])
    .join("\n");

  // write one time
  var stream = fs.writeFileSync(FILEPATH, fileContents, handleError);

  // use this to spot check the output
  console.log(entriesByOwner);
};

const runMain = async () => {
  try {
    await main();
    process.exit(0);
  } catch (error) {
    console.log(error);
    process.exit(1);
  }
};

runMain();