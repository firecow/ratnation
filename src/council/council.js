import http from "http";
import findmyway from "find-my-way";
import {Provisioner} from "./provisioner.js";
import putling from "./put-ling.js";
import getState from "./get-state.js";
import putKing from "./put-king.js";

export const command = "council";
export const description = "Start council";

export async function handler(argv) {
    const state = {
        revision: 0,
        services: [],
        kings: [],
        lings: [],
    };

    const provisioner = new Provisioner({state});

    const router = findmyway({
        defaultRoute: (req, res) => {
            res.statusCode = 404;
            res.end();
        }
    });

    router.on("GET", "/state", (req, res) => getState(req, res, state));
    router.on("PUT", "/ling", (req, res) => putling(req, res, state));
    router.on("PUT", "/king", (req, res) => putKing(req, res, state));

    const server = http.createServer((req, res) => router.lookup(req, res));
    server.listen(argv["port"]);
    await new Promise(resolve => server.once("listening", resolve));
    console.log("msg=\"council ready\" service_type=ratcouncil");

    provisioner.start();
}

export function builder(yargs) {
    yargs.options("port", {
        type: "number",
        description: "Webserver listening port",
        default: "8080",
    });
}
