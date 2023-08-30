import "source-map-support/register.js";
import assert from "assert";
import yargs from "yargs";
import chalk from "chalk-template";
import * as councilCmd from "./council/council.js";
import * as requesterCmd from "./debug/requester.js";
import * as kingCmd from "./king/king.js";
import * as lingCmd from "./ling/ling.js";

process.on("uncaughtException", (err) => {
    if (err instanceof assert.AssertionError) {
        console.error(chalk`{red ${err.message}}`);
    } else {
        console.error(err.message, err.stack?.split("\n").slice(0, 2).join("\n"));
    }
    process.exit(1);
});

const terminalWidth = yargs().terminalWidth();
await yargs(process.argv.slice(2))
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
