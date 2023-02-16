import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.mjs";
import createServer from "./council-server.mjs";

export const command = "council";
export const description = "Start council";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const server = createServer();
    server.listen(args.port);
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
