import {ArgumentsCamelCase, Argv} from "yargs";
import {Logger} from "../logger.js";
import createServer from "../council-server.js";
import {initCouncilShutdownHandlers} from "../shutdown/council-shutdown.js";
import {CouncilProvisioner} from "../council-provisioner.js";
import {CouncilStateCleaner} from "../tickers/coucil-state-cleaner.js";

export const command = "council";
export const description = "Start council";

export async function handler (args: ArgumentsCamelCase) {
    const logger = new Logger();

    const state = {revision: 0, services: [], kings: [], lings: []};
    const provisioner = new CouncilProvisioner({logger});
    const {httpServer, socketIo} = createServer({provisioner, state});
    const cleaner = new CouncilStateCleaner({state, logger, socketIo});

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
