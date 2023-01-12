import yargs from "yargs";
import assert from "assert";
import * as kingCmd from "./king/king.js";
import * as lingCmd from "./ling/ling.js";
import * as councilCmd from "./council/council.js";

Array.prototype.random = function() {
    return this[Math.floor((Math.random() * this.length))];
};

process.on("uncaughtException", (err) => {
    if (err instanceof assert.AssertionError) {
        console.error(err.message);
    } else {
        console.log(err.message, err.stack.split("\n").slice(0, 2).join("\n"));
    }
    process.exit(1);
});

const terminalWidth = yargs().terminalWidth();
const y = yargs(process.argv.slice(2))
    .command(councilCmd)
    .command(kingCmd)
    .command(lingCmd)
    .demandCommand()
    .fail((msg, err) => {
        if (!err) throw new assert.AssertionError({message: msg});
    })
    .wrap(terminalWidth)
    .strict(true);
y.parse();
