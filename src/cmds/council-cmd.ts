import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.js";
import createServer from "../council-server.js";
import {initCouncilShutdownHandlers} from "../shutdown/council-shutdown.js";

export const command = "council";
export const description = "Start council";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();
    const {httpServer, cleaner} = createServer();
    httpServer.listen(args.port);
    await new Promise(resolve => httpServer.once("listening", resolve));
    cleaner.start();
    initCouncilShutdownHandlers({logger, httpServer, cleaner});
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
