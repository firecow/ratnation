import yargs from "yargs";
import assert from "assert";
import * as kingCmd from "./king/king.mjs";
import * as lingCmd from "./ling/ling.mjs";
import * as councilCmd from "./council/council.mjs";
import * as requesterCmd from "./debug/requester.mjs";

process.on("uncaughtException", (err) => {
    if (err instanceof assert.AssertionError) {
        console.error(err.message);
    } else {
        console.log(err.message, err.stack?.split("\n").slice(0, 2).join("\n"));
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
