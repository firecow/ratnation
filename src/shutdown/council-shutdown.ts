import {Server} from "http";
import {Logger} from "../logger.js";
import {CouncilStateCleaner} from "../tickers/coucil-state-cleaner.js";

interface CouncilShutdownHandlersOpts {
    logger: Logger;
    httpServer: Server;
    cleaner: CouncilStateCleaner;
}

let shuttingDown = false;

export function initCouncilShutdownHandlers ({logger, httpServer, cleaner}: CouncilShutdownHandlersOpts) {
    const listener = () => {
        if (shuttingDown) return;
        shuttingDown = true;
        logger.info("Shutdown sequence initiated", {"service.type": "ratcouncil"});
        cleaner.stop();
        httpServer.close();
    };
    process.on("SIGINT", listener);
    process.on("SIGTERM", listener);
}
