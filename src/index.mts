import assert from "assert";
import "source-map-support/register.js";
import yargs from "yargs";
import * as councilCmd from "./council/council.mjs";
import * as requesterCmd from "./debug/requester.mjs";
import * as kingCmd from "./king/king.mjs";
import * as lingCmd from "./ling/ling.mjs";

process.on("uncaughtException", (err) => {
    if (err instanceof assert.AssertionError) {
        console.error(`\u001b[31m${err.message}\u001b[0m`);
    } else {
        console.error(err.message, err.stack?.split("\n").slice(0, 2).join("\n"));
    }
    process.exit(1);
});

const terminalWidth = yargs().terminalWidth();
// eslint-disable-next-line @typescript-eslint/no-floating-promises
yargs(process.argv.slice(2))
    .command(councilCmd)
    .command(kingCmd)
    .command(lingCmd)
    .command(requesterCmd)
    .demandCommand()
    .fail((msg, err) => {
        if (!err) throw new assert.AssertionError({message: msg});
    })
    .wrap(terminalWidth)
    .strict(true)
    .parse();
