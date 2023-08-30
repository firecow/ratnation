import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.js";
import createServer from "./council-server.js";
import {initCouncilShutdownHandlers} from "./council-shutdown.js";

export const command = "council";
export const description = "Start council";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const {server, cleaner} = createServer();
    server.listen(args.port);
    await new Promise(resolve => server.once("listening", resolve));
    cleaner.start();
    initCouncilShutdownHandlers({logger, server, cleaner});
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
