import yargs from "yargs";
import assert from "assert";
import * as kingCmd from "./king/king.js";
import * as underlingCmd from "./underling/underling.js";
import * as councilCmd from "./council/council.js";

Array.prototype.random = function() {
    return this[Math.floor((Math.random() * this.length))];
};

process.on("uncaughtException", (err) => {
    if (err instanceof assert.AssertionError) {
        console.error(err.message);
    } else {
        console.log(err.message);
    }
    process.exit(1);
});

const terminalWidth = yargs().terminalWidth();
const y = yargs(process.argv.slice(2))
    .command(councilCmd)
    .command(kingCmd)
    .command(underlingCmd)
    .demandCommand()
    .fail((msg, err) => {
        if (!err) throw new assert.AssertionError({message: msg});
    })
    .wrap(terminalWidth)
    .strict(true);
y.parse();
