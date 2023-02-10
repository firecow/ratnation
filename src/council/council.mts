import http from "http";
import findmyway from "find-my-way";
import {Provisioner} from "./provisioner.mjs";
import putLing from "./put-ling.mjs";
import getState from "./get-state.mjs";
import putKing from "./put-king.mjs";
import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.mjs";

export const command = "council";
export const description = "Start council";

export async function handler (argv: ArgumentsCamelCase) {
    const logger = new Logger();
    const state = {
        revision: 0,
        services: [],
        kings: [],
        lings: [],
    };

    const provisioner = new Provisioner({logger, state});

    const router = findmyway({
        defaultRoute: (req, res) => {
            res.statusCode = 404;
            res.end();
        },
    });

    router.on("GET", "/state", async (req, res) => getState(logger, req, res, state));
    router.on("PUT", "/ling", async (req, res) => putLing(logger, req, res, state, provisioner));
    router.on("PUT", "/king", async (req, res) => putKing(logger, req, res, state, provisioner));

    const server = http.createServer((req, res) => router.lookup(req, res));
    server.listen(argv.port);
    await new Promise(resolve => server.once("listening", resolve));
    logger.info("Ready", {"service.type": "ratcouncil"});
}

export function builder (yargs: Argv) {
    yargs.options("port", {
        type: "number",
        description: "Webserver listening port",
        default: "8080",
    });
    return yargs;
}
